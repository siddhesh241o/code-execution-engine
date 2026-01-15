package redis

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type RedisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{
		client: client,
		key:    "judge:queue:pending",
	}
}

func (rq *RedisQueue) Push(ctx context.Context, job domain.ExecutionRequest) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return rq.client.RPush(ctx, rq.key, data).Err()
}

func (rq *RedisQueue) Pop(ctx context.Context) (domain.ExecutionRequest, error) {
	data, err := rq.client.BLPop(ctx, 0, rq.key).Result()
	if err != nil {
		return domain.ExecutionRequest{}, err
	}
	var job domain.ExecutionRequest
	err = json.Unmarshal([]byte(data[1]), &job)
	return job, err
}
