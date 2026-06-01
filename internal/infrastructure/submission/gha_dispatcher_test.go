package submission

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/redis"
)

func TestDispatch(t *testing.T) {
	ghc := GithubActionsConfig{
		Owner:          os.Getenv("GITHUB_OWNER"),
		Repo:           os.Getenv("GITHUB_REPO"),
		Workflow:       os.Getenv("GITHUB_WORKFLOW"),
		Ref:            os.Getenv("GITHUB_REF"),
		Token:          os.Getenv("GITHUB_TOKEN"),
		BaseURL:        "https://api.github.com",
		FetchURL:       os.Getenv("FETCH_URL"),
		CallbackURL:    os.Getenv("CALLBACK_URL"),
		FetchSecret:    os.Getenv("FETCH_SECRET"),
		CallbackSecret: os.Getenv("CALLBACK_SECRET"),
	}

	if ghc.Owner == "" || ghc.Repo == "" || ghc.Workflow == "" || ghc.Ref == "" || ghc.Token == "" {
		t.Fatalf("missing github config")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb, err := redis.NewRedisClient(redisAddr)
	if err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}
	store := redis.NewRedisJobInfoStore(rdb, 10*time.Minute)

	dispatcher, err := NewGitHubActionsSubmissionDispatcher(ghc, nil)
	if err != nil {
		t.Fatalf("failed to create dispatcher: %v", err)
	}

	jobID := "test-" + uuid.New().String()
	req := domain.ExecutionRequest{
		ID:       jobID,
		Language: "python",
		Code:     "import sys; print('Hello from GHA Test!'); print(sys.version)",
	}

	if err := store.Set(context.Background(), req); err != nil {
		t.Fatalf("failed to store job in redis: %v", err)
	}
	t.Logf("Job %s stored in Redis", jobID)

	if err := dispatcher.Dispatch(context.Background(), req); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	t.Logf("Successfully triggered GHA for Job: %s", jobID)
	t.Logf("Check your GitHub Actions tab and then poll: https://millesimally-unmingled-arnetta.ngrok-free.dev/api/jobs/%s/result", jobID)
}
