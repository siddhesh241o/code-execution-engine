package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type RedisJobInfoStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisJobInfoStore(client *redis.Client, ttl time.Duration) *RedisJobInfoStore {
	return &RedisJobInfoStore{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisJobInfoStore) Set(ctx context.Context, res domain.ExecutionRequest) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("job:%s", res.ID)
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *RedisJobInfoStore) Get(ctx context.Context, id string) (*domain.ExecutionRequest, error) {
	key := fmt.Sprintf("job:%s", id)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get job failed: %v", err)
	}
	var resp domain.ExecutionRequest
	err = json.Unmarshal([]byte(data), &resp)
	return &resp, err
}
