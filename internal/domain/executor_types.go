package domain

import (
	"context"
	"time"
)

type Status int 

const(
	StatusSuccessfullyExecuted Status = iota
	StatusError
	StatusTLE 
	StatusMLE
	StatusSystemError
)

func(s Status) String() string {
	return [...]string{
		"Successfully Executed",
		"Error",
		"Time Limit Exceeded", 
		"Memory Limit",
		"System Error",
	}[s]
}

type ExecutionRequest struct {
	Code string
	Language string
}

type ExecutionResponse struct {
	Stdout string 
	Stderr string
	Duration time.Duration
	Status 
}

type CodeExecutor interface {
	Execute(ctx context.Context, req ExecutionRequest) (*ExecutionRequest, error)
}

