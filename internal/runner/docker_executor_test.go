package runner

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/moby/moby/client"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

//  Command being timed: "./app"
//     User time (seconds): 0.00
//     System time (seconds): 0.00
//     Percent of CPU this job got: 25%
//     Elapsed (wall clock) time (h:mm:ss or m:ss): 0:00.00
//     Average shared text size (kbytes): 0
//     Average unshared data size (kbytes): 0
//     Average stack size (kbytes): 0
//     Average total size (kbytes): 0
//     Maximum resident set size (kbytes): 2560
//     Average resident set size (kbytes): 0
//     Major (requiring I/O) page faults: 19
//     Minor (reclaiming a frame) page faults: 93
//     Voluntary context switches: 14
//     Involuntary context switches: 0
//     Swaps: 0
//     File system inputs: 3264
//     File system outputs: 0
//     Socket messages sent: 0
//     Socket messages received: 0
//     Signals delivered: 0
//     Page size (bytes): 4096
//     Exit status: 0

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
		containsStderr string
	}{
		{
			name: "Python Success with Input",
			req: domain.ExecutionRequest{
				Language: "python",
				Code:     `name = input(); print(f"Hello {name}")`,
				Input:    "Siddhesh",
			},
			expectedStatus: domain.StatusSuccessfullyExecuted,
			containsStdout: "Hello Siddhesh",
		},
		{
			name: "Python Runtime Error (Division by Zero)",
			req: domain.ExecutionRequest{
				Language: "python",
				Code:     `print(1/0)`,
			},
			expectedStatus: domain.StatusRuntimeError,
		},
		{
			name: "Python Memory Limit Exceeded (MLE)",
			req: domain.ExecutionRequest{
				Language: "python",
				Code:     `a = [1] * (20 * 1024 * 1024) # Attempt to allocate large list`,
			},
			expectedStatus: domain.StatusMLE,
		},
		{
			name: "C++ Compilation Error",
			req: domain.ExecutionRequest{
				Language: "c++",
				Code:     `int main() { std::cout << "Missing include" }`,
			},
			expectedStatus: domain.StatusCompileError,
		},
		{
			name: "C++ Time Limit Exceeded (TLE)",
			req: domain.ExecutionRequest{
				Language: "c++",
				Code:     `int main() { while(true); return 0; }`,
			},
			expectedStatus: domain.StatusTLE,
		},
		{
			name: "Java Successful with Input",
			req: domain.ExecutionRequest{
				Language: "java",
				Code: `
                import java.util.Scanner;
                public class Main {
                    public static void main(String[] args) {
                        Scanner sc = new Scanner(System.in);
                        if (sc.hasNext()) {
                            System.out.println("Hello " + sc.next());
                        }
                    }
                }`,
				Input: "Siddhesh",
			},
			expectedStatus: domain.StatusSuccessfullyExecuted,
			containsStdout: "Hello Siddhesh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			res, err := executor.Execute(ctx, tt.req)

			if err != nil {
				t.Fatalf("Execute returned unexpected error: %v", err)
			}

			fmt.Printf("stderr:%v, stdout:%v", res.Stderr, res.Stdout)
			if res.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, res.Status)
			}
			if tt.containsStdout != "" && !strings.Contains(res.Stdout, tt.containsStdout) {
				t.Errorf("Expected stdout to contain %q, got %q", tt.containsStdout, res.Stdout)
			}

			t.Logf("--- %s ---", tt.name)
			t.Logf("Status: %v", res.Status)
			t.Logf("Stdout: %q", res.Stdout)
			t.Logf("Stderr: %q", res.Stderr)
			t.Logf("Duration: %vms", res.Duration.Milliseconds())
		})
	}
}
