package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type DockerExecutor struct {
	cli *client.Client
	fm *FileManager
}

func NewDockerExecutor(cli *client.Client, fm *FileManager) *DockerExecutor {
	return &DockerExecutor{
		cli: cli,
		fm: fm,
	}
}

func (de *DockerExecutor) ensureImage(ctx context.Context, image string) error {
	_, err := de.cli.ImageInspect(ctx, image)
	if err == nil {
		return nil
	}
	pullCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	fmt.Printf("Image not found, Pulling the image\n")
	
	reader, err := de.cli.ImagePull(pullCtx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull failed: %v", err)
	}
	defer reader.Close()

	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return fmt.Errorf("stream error, pull failed: %v", err)
	}
	fmt.Printf("image pulled successfully")
	return nil
}

func (de *DockerExecutor) calculateDuration(inspect client.ContainerInspectResult) time.Duration{
	start, _ := time.Parse(time.RFC3339Nano, inspect.Container.State.StartedAt)
	end, _ := time.Parse(time.RFC3339Nano, inspect.Container.State.FinishedAt)
	duration := end.Sub(start)
	if duration < 0 {
		return 0
	}
	return duration
}

func (de *DockerExecutor) determineStatus(inspect client.ContainerInspectResult, ctxErr error) domain.Status {
	if ctxErr != nil {
		return domain.StatusTLE
	}
	state := inspect.Container.State
	if state.OOMKilled || state.ExitCode == 137 {
		return domain.StatusMLE
	}
	if state.ExitCode == 0 {
		return domain.StatusSuccessfullyExecuted
	}
	return domain.StatusError
}

func (de *DockerExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (*domain.ExecutionResponse, error) {
	languageConfig, err := GetLanguageConfig(req.Language)
	if err != nil {
		return nil, err
	}
	err = de.ensureImage(ctx, languageConfig.Image)
	if err != nil {
		return nil, fmt.Errorf("image does not exist: %v", err)
	}

	workdir, cleanup, err := de.fm.CreateWorkspace(req.Code, languageConfig.SourceFile)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	
	containerConfig := container.Config{
		Image: languageConfig.Image,
		Cmd: languageConfig.Command,
		WorkingDir: "/code",
		Tty: false,
		NetworkDisabled: true,
	}

	limit := int64(32)
	hostConfig := container.HostConfig{
		Binds: []string{workdir +":/code"},
		Resources: container.Resources{
			Memory: 128*1024*1024,
			NanoCPUs: 500000000,
			PidsLimit: &limit,
		},
	}

	resp, err := de.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &containerConfig,
		HostConfig: &hostConfig,
		NetworkingConfig:nil,
		Platform: nil,
		Name: "",
	})
	if err != nil {
		return nil, fmt.Errorf("container creation failed: %v", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		de.cli.ContainerRemove(cleanupCtx, resp.ID,client.ContainerRemoveOptions{})
	}()

	runCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := de.cli.ContainerStart(runCtx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}
	wait := de.cli.ContainerWait(runCtx, resp.ID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})

	select {
	case err = <- wait.Error:
		if err != nil {
			return nil, fmt.Errorf("failed waiting for container: %v", err)
		}
	case <- runCtx.Done():
		de.cli.ContainerKill(context.Background(), resp.ID, client.ContainerKillOptions{Signal: "SIGKILL"})
	case <- wait.Result:
	}

	inspect, err := de.cli.ContainerInspect(context.Background(), resp.ID, client.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	logs, err := de.cli.ContainerLogs(context.Background(), resp.ID, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get exection logs: %v", err)
	}

	_, _ = stdcopy.StdCopy(&outBuf, &errBuf, logs)
	logs.Close()
	return &domain.ExecutionResponse{
		Stdout: outBuf.String(),
		Stderr: errBuf.String(),
		Duration : de.calculateDuration(inspect),
		Status : de.determineStatus(inspect, runCtx.Err()),
	}, nil
}