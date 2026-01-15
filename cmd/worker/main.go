package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/moby/moby/client"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/redis"
	infraRedis "github.com/siddhesh241o/code-execution-engine/internal/infrastructure/redis"
	"github.com/siddhesh241o/code-execution-engine/internal/runner"
)

const (
	WorkerCount = 4
)

func workerLoop(workerId int, queue *infraRedis.RedisQueue, store *infraRedis.RedisResultStore, executor *runner.DockerExecutor) {
	ctx := context.Background()
	for {
		job, err := queue.Pop(ctx)
		if err != nil {
			log.Printf("Worker_%d queue error:%v", workerId, err)
		}
		result, err := executor.Execute(ctx, job)
		if err != nil {
			log.Printf("Worker_%d failed job execution %s: %v", workerId, job.ID, err)
			store.Set(ctx, domain.ExecutionResponse{
				ID:     job.ID,
				Status: domain.StatusSystemError.String(),
			})
			continue
		}
		result.ID = job.ID
		if err = store.Set(ctx, *result); err != nil {
			log.Printf("Worker_%d failed to save result for job_%s: %v", workerId, job.ID, err)
		}
	}
}

func main() {
	rdb, err := infraRedis.NewRedisClient("localhost:6379")
	if err != nil {
		log.Fatalf("redis client creation failed: %v", err)
	}
	queue := redis.NewRedisQueue(rdb)
	store := redis.NewRedisResultStore(rdb, time.Minute*5)
	dc, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatalf("docker client creation failed: %v", err)
	}
	fm := runner.NewFileManager()
	executor := runner.NewDockerExecutor(dc, fm)

	log.Printf("Spawning workers\n")
	var wg sync.WaitGroup
	for i := range WorkerCount {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			workerLoop(workerId, queue, store, executor)
		}(i)
	}
	wg.Wait()
}
