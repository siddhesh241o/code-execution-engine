package runner

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	_ "embed"

	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

//go:embed images/cpp.Dockerfile
var cppDockerfile string

//go:embed images/java.Dockerfile
var javaDockerfile string

//go:embed images/python.Dockerfile
var pythonDockerfile string

type DockerExecutor struct {
	cli *client.Client
	fm  *FileManager
}

type runResult struct {
	Stdout   string
	Stderr   string
	Inspect  client.ContainerInspectResult
	ctxError error
}

func NewDockerExecutor(cli *client.Client, fm *FileManager) *DockerExecutor {
	de := DockerExecutor{
		cli: cli,
		fm:  fm,
	}
	de.ensureCustomImages()
	return &de
}

func (de *DockerExecutor) ensureCustomImages() {
	images := map[string]string{
		"engine-cpp":    cppDockerfile,
		"engine-java":   javaDockerfile,
		"engine-python": pythonDockerfile,
	}
	ctx := context.Background()
	for name, content := range images {
		_, err := de.cli.ImageInspect(ctx, name)
		if err == nil {
			continue
		}
		_ = de.buildCustomImage(ctx, name, content)
	}
}

func (de *DockerExecutor) buildCustomImage(ctx context.Context, tagName string, dockerfile string) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfile)),
	}
	err := tw.WriteHeader(tarHeader)
	if err != nil {
		return err
	}
	_, err = tw.Write([]byte(dockerfile))
	if err != nil {
		return err
	}
	tw.Close()
	res, err := de.cli.ImageBuild(ctx, buf, client.ImageBuildOptions{
		Tags:        []string{tagName},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
	})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(os.Stdout, res.Body)
	return nil
}

func (de *DockerExecutor) determineStatus(inspect client.ContainerInspectResult, ctxErr error, memoryKB int64) domain.Status {
	if ctxErr != nil {
		return domain.StatusTLE
	}
	state := inspect.Container.State
	if state.OOMKilled || memoryKB > 128*1024 {
		return domain.StatusMLE
	}
	if state.ExitCode == 0 {
		return domain.StatusSuccessfullyExecuted
	}
	return domain.StatusRuntimeError
}

func (de *DockerExecutor) parseGNUTimeString(content string) (time.Duration, int64) {
	var userTime, systemTime float64
	var maxMemoryKB int64
	regUser := regexp.MustCompile(`User time \(seconds\): ([\d.]+)`)
	regSystem := regexp.MustCompile(`System time \(seconds\): ([\d.]+)`)
	regMem := regexp.MustCompile(`Maximum resident set size \(kbytes\): (\d+)`)
	if match := regUser.FindStringSubmatch(content); len(match) > 1 {
		userTime, _ = strconv.ParseFloat(match[1], 64)
	}
	if match := regSystem.FindStringSubmatch(content); len(match) > 1 {
		systemTime, _ = strconv.ParseFloat(match[1], 64)
	}
	if match := regMem.FindStringSubmatch(content); len(match) > 1 {
		maxMemoryKB, _ = strconv.ParseInt(match[1], 10, 64)
	}
	totalSeconds := userTime + systemTime
	return time.Duration(totalSeconds * float64(time.Second)), maxMemoryKB
}

func (de *DockerExecutor) runInContainer(ctx context.Context, languageConfig LanguageConfig, workdir string, input string) (*runResult, error) {
	containerConfig := container.Config{
		Image:           languageConfig.Image,
		Cmd:             languageConfig.Command,
		WorkingDir:      "/code",
		OpenStdin:       true,
		StdinOnce:       true,
		Tty:             false,
		NetworkDisabled: true,
	}

	limit := int64(32)
	hostConfig := container.HostConfig{
		Binds: []string{workdir + ":/code"},
		Resources: container.Resources{
			Memory:     128 * 1024 * 1024,
			NanoCPUs:   500000000,
			MemorySwap: 128 * 1024 * 1024,
			PidsLimit:  &limit,
		},
	}

	resp, err := de.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:           &containerConfig,
		HostConfig:       &hostConfig,
		NetworkingConfig: nil,
		Platform:         nil,
		Name:             "",
	})
	if err != nil {
		return nil, fmt.Errorf("container creation failed: %v", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = de.cli.ContainerRemove(cleanupCtx, resp.ID, client.ContainerRemoveOptions{})
	}()

	waiter, err := de.cli.ContainerAttach(ctx, resp.ID, client.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stderr: true,
		Stdout: true,
	})
	if err != nil {
		return nil, fmt.Errorf("container failed to attach: %v", err)
	}
	defer waiter.Close()

	go func() {
		if input != "" {
			_, _ = io.WriteString(waiter.Conn, input)
		}
		waiter.CloseWrite()
	}()

	runCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := de.cli.ContainerStart(runCtx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}
	wait := de.cli.ContainerWait(runCtx, resp.ID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})

	select {
	case err = <-wait.Error:
		if err != nil {
			return nil, fmt.Errorf("failed waiting for container: %v", err)
		}
	case <-runCtx.Done():
		de.cli.ContainerKill(context.Background(), resp.ID, client.ContainerKillOptions{Signal: "SIGKILL"})
	case <-wait.Result:
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
	return &runResult{
		Stdout:   outBuf.String(),
		Stderr:   errBuf.String(),
		Inspect:  inspect,
		ctxError: runCtx.Err(),
	}, nil
}

func (de *DockerExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (*domain.ExecutionResponse, error) {
	languageConfig, err := GetLanguageConfig(req.Language)
	if err != nil {
		return nil, err
	}
	_, err = de.cli.ImageInspect(ctx, languageConfig.Image)
	if err != nil {
		return nil, fmt.Errorf("image does not exist: %v", err)
	}

	workdir, cleanup, err := de.fm.CreateWorkspace(req.Code, languageConfig.SourceFile)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	result, err := de.runInContainer(ctx, languageConfig, workdir, req.Input)
	if err != nil {
		return nil, fmt.Errorf("execution failed %v", err)
	}

	metricsPath := filepath.Join(workdir, "metrics.txt")
	metricsData, err := os.ReadFile(metricsPath)
	if err != nil {
		return &domain.ExecutionResponse{
			Stdout: result.Stdout,
			Stderr: result.Stderr,
			Status: domain.StatusCompileError,
		}, nil
	}

	duration, memoryKB := de.parseGNUTimeString(string(metricsData))

	return &domain.ExecutionResponse{
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Duration: duration,
		Status:   de.determineStatus(result.Inspect, result.ctxError, memoryKB),
	}, nil
}
