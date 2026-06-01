package submission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type GithubActionsConfig struct {
	Owner          string
	Repo           string
	Workflow       string
	Ref            string
	Token          string
	BaseURL        string
	FetchURL       string
	CallbackURL    string
	FetchSecret    string
	CallbackSecret string
}

type GithubActionsDispatcher struct {
	cfg        GithubActionsConfig
	httpClient *http.Client
}

func NewGitHubActionsSubmissionDispatcher(cfg GithubActionsConfig, client *http.Client) (*GithubActionsDispatcher, error) {

	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &GithubActionsDispatcher{
		cfg:        cfg,
		httpClient: client,
	}, nil
}

func (d *GithubActionsDispatcher) Dispatch(ctx context.Context, req domain.ExecutionRequest) error {
	type dispatchPayload struct {
		Ref    string            `json:"ref"`
		Inputs map[string]string `json:"inputs"`
	}

	payload := dispatchPayload{
		Ref: d.cfg.Ref,
		Inputs: map[string]string{
			"job_id":          req.ID,
			"fetch_url":       d.cfg.FetchURL,
			"fetch_secret":    d.cfg.FetchSecret,
			"callback_url":    fmt.Sprintf(d.cfg.CallbackURL, req.ID),
			"callback_secret": d.cfg.CallbackSecret,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal failed on workflow payload: %w", err)
	}
	url := fmt.Sprintf(
		"%s/repos/%s/%s/actions/workflows/%s/dispatches",
		d.cfg.BaseURL,
		d.cfg.Owner,
		d.cfg.Repo,
		d.cfg.Workflow,
	)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("http request to github failed: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.cfg.Token)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("github dispatch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("github dispatch failed: status%d body=%s", resp.StatusCode, string(respBody))
	}
	return nil
}
