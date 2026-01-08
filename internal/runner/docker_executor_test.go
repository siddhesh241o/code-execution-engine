package runner

import (
	"context"
	"testing"
	"strings"

	"github.com/moby/moby/client"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

func TestEnsureImage_Integration(t *testing.T) {
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		t.Fatalf("unable to create docker client: %v", err)
	}
	defer apiClient.Close()
	executor := NewDockerExecutor(apiClient, NewFileManager())
	ctx := context.Background()
	image := "alpine:latest"

	_, _ = apiClient.ImageRemove(ctx, image, client.ImageRemoveOptions{})
	err = executor.ensureImage(ctx, image)
	if err != nil {
		t.Errorf("ensure image failed to pull %v", err)
	}
	_, err = apiClient.ImageInspect(ctx, image)
	if err != nil {
		t.Errorf("image should exist after pull but got %v", err)
	}
	_, err = apiClient.ImageRemove(ctx, image, client.ImageRemoveOptions{})
	if err != nil {
		t.Logf("Cleanup warning: could not remove image: %v", err)
	}
}

func TestDockerExecutor_Execute_Integration(t *testing.T) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		t.Fatalf("Failed to connect to Docker: %v", err)
	}
	defer cli.Close()

	fm := NewFileManager()
	executor := NewDockerExecutor(cli, fm)

	tests := []struct {
		name           string
		req            domain.ExecutionRequest
		expectedStatus domain.Status
		containsStdout string
	}{
		{
			name: "Python Success",
			req: domain.ExecutionRequest{
				Language: "python",
				Code:     `print("Hello World")`,
			},
			expectedStatus: domain.StatusSuccessfullyExecuted,
			containsStdout: "Hello World",
		},
		{
			name: "Python Syntax Error",
			req: domain.ExecutionRequest{
				Language: "python",
				Code:     `print("Hello"`, 
			},
			expectedStatus: domain.StatusError,
		},
		{
			name: "Python Timeout (TLE)",
			req: domain.ExecutionRequest{
				Language: "python",
				Code:     "import time\nwhile True: time.sleep(1)", 
			},
			expectedStatus: domain.StatusTLE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			res, err := executor.Execute(ctx, tt.req)

			if err != nil {
				t.Fatalf("Execute returned unexpected error: %v", err)
			}

			if res.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, res.Status)
			}

			if tt.containsStdout != "" && !strings.Contains(res.Stdout, tt.containsStdout) {
				t.Errorf("Expected stdout to contain %q, got %q", tt.containsStdout, res.Stdout)
			}
			
			t.Logf("Result - Status: %v, Duration: %v", res.Status, res.Duration)
		})
	}
}