package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type RedisResultStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisResultStore(client *redis.Client, ttl time.Duration) *RedisResultStore {
	return &RedisResultStore{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisResultStore) Set(ctx context.Context, res domain.ExecutionResponse) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("result:%s", res.ID)
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *RedisResultStore) Get(ctx context.Context, id string) (*domain.ExecutionResponse, error) {
	key := fmt.Sprintf("result:%s", id)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get execution result failed: %v", err)
	}
	var resp domain.ExecutionResponse
	err = json.Unmarshal([]byte(data), &resp)
	return &resp, err
}
