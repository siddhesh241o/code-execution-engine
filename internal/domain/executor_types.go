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
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required"`
	Input    string `json:"input"`
}

type ExecutionResponse struct {
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	Duration time.Duration `json:"time_ms"`
	Memory   int64         `json:"memory_kb"`
	Status   `json:"status"`
}

type CodeExecutor interface {
	Execute(ctx context.Context, req ExecutionRequest) (*ExecutionRequest, error)
}
