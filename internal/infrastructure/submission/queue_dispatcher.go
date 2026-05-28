package submission

import (
	"context"
	"fmt"

	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type QueueDispatcher struct {
	Queue domain.JobQueue
}

func NewQueueSubmissionDispatcher(queue domain.JobQueue) *QueueDispatcher {
	return &QueueDispatcher{
		Queue: queue,
	}
}

func (q *QueueDispatcher) Dispatch(ctx context.Context, req domain.ExecutionRequest) error {
	err := q.Queue.Push(ctx, req)
	if err != nil {
		return fmt.Errorf("queue push failed: %v", err)
	}
	return nil
}
