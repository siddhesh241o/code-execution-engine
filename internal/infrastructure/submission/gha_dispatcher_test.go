package submission

import (
	"context"
	"os"
	"testing"

	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

func TestDispatch(t *testing.T) {
	ghc := GithubActionsConfig{
        Owner:          os.Getenv("GITHUB_OWNER"),
        Repo:           os.Getenv("GITHUB_REPO"),
        Workflow:       os.Getenv("GITHUB_WORKFLOW"),
        Ref:            os.Getenv("GITHUB_REF"),
        Token:          os.Getenv("GITHUB_TOKEN"),
        FetchURL:       os.Getenv("FETCH_URL"),       
        CallbackURL:    os.Getenv("CALLBACK_URL"),    
        FetchSecret:    os.Getenv("FETCH_SECRET"),
        CallbackSecret: os.Getenv("CALLBACK_SECRET"),
    }

	//t.Logf("owner=%q repo=%q workflow=%q ref=%q", ghc.Owner, ghc.Repo, ghc.Workflow, ghc.Ref)

    if ghc.Owner == "" || ghc.Repo == "" || ghc.Workflow == "" || ghc.Ref == "" || ghc.Token == "" {
        t.Fatalf("missing github config")
    }
	
	dispatcher, err := NewGitHubActionsSubmissionDispatcher(ghc, nil)
	if err != nil {
		t.Fatalf("failed to create dispatcher: %v", err)
	}
	req := domain.ExecutionRequest{
		ID: "test-job-1",
		Language: "python",
		Code: "print('hello')",
	}

	if err := dispatcher.Dispatch(context.Background(), req); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
}
