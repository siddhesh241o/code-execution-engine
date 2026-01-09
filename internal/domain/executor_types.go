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
	Code     string
	Language string
	Input    string
}

type ExecutionResponse struct {
	Stdout   string
	Stderr   string
	Duration time.Duration
	Status
}

type CodeExecutor interface {
	Execute(ctx context.Context, req ExecutionRequest) (*ExecutionRequest, error)
}
