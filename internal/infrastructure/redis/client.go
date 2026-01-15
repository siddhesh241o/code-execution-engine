package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("redis client creation failed: %v", err)
	}
	return rdb, nil
}
