package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
	infraRedis "github.com/siddhesh241o/code-execution-engine/internal/infrastructure/redis"
	"github.com/siddhesh241o/code-execution-engine/internal/observability"
	"github.com/siddhesh241o/code-execution-engine/internal/runner"
)

var (
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
			observability.RecordExecution(job.Language, domain.StatusSystemError.String(), 0)
			store.Set(ctx, domain.ExecutionResponse{
				ID:     job.ID,
				Status: domain.StatusSystemError.String(),
			})
			continue
		}
		result.ID = job.ID
		observability.RecordExecution(job.Language, result.Status, result.Duration)
		if err = store.Set(ctx, *result); err != nil {
			log.Printf("Worker_%d failed to save result for job_%s: %v", workerId, job.ID, err)
			continue
		}
		log.Printf("Worker_%d completed job_%s", workerId, job.ID)
	}
}

func main() {
	_ = godotenv.Load()
	redisAddr := os.Getenv("REDDIS_ADDR")
	ttlStr := os.Getenv("RESULT_TTL")
	if wc := os.Getenv("WORKER_COUNT"); wc != "" {
		if n, err := strconv.Atoi(wc); err == nil && n > 0 {
			WorkerCount = n
		}
	}
	ttl, err := time.ParseDuration(ttlStr)

	if err != nil {
		log.Fatalf("invalid RESULT_TTL: %v", err)
	}
	rdb, err := infraRedis.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatalf("redis client creation failed: %v", err)
	}
	queue := infraRedis.NewRedisQueue(rdb)
	store := infraRedis.NewRedisResultStore(rdb, ttl)
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
