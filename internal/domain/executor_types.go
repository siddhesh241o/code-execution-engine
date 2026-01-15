package domain

import (
	"context"
	"time"
)

type Status int

const (
	StatusSuccessfullyExecuted Status = iota
	StatusCompileError
	StatusRuntimeError
	StatusTLE
	StatusMLE
	StatusSystemError
)

func (s Status) String() string {
	return [...]string{
		"Successfully Executed",
		"Compilation Error",
		"Runtime Error",
		"Time Limit Exceeded",
		"Memory Limit",
		"System Error",
	}[s]
}

type ExecutionRequest struct {
	ID        string    `json:"id"`
	Code      string    `json:"code" binding:"required"`
	Language  string    `json:"language" binding:"required"`
	Input     string    `json:"input"`
	CreatedAt time.Time `json:"created_at"`
}

type ExecutionResponse struct {
	ID       string        `json:"id"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	Duration time.Duration `json:"time_ms"`
	Memory   int64         `json:"memory_kb"`
	Status   string        `json:"status"`
}

type CodeExecutor interface {
	Execute(ctx context.Context, req ExecutionRequest) (*ExecutionRequest, error)
}

type QueueProvider interface {
	Push(ctx context.Context, e ExecutionRequest) error
	Pop(ctx context.Context) (ExecutionRequest, error)
}

type ResultStore interface {
	Set(ctx context.Context, res ExecutionResponse) error
	Get(ctx context.Context, id string) (*ExecutionResponse, error)
}
