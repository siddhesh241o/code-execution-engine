package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type InMemoryProvider struct {
	mu             sync.RWMutex
	jobInfos       map[string]domain.ExecutionRequest
	jobResults     map[string]domain.ExecutionResponse
	jobQueue       chan domain.ExecutionRequest
	ttl            time.Duration
	expirationKeys map[string]time.Time
}

func NewInMemoryProvider(ttl time.Duration) *InMemoryProvider {
	p := &InMemoryProvider{
		jobInfos:       make(map[string]domain.ExecutionRequest),
		jobResults:     make(map[string]domain.ExecutionResponse),
		jobQueue:       make(chan domain.ExecutionRequest, 100),
		ttl:            ttl,
		expirationKeys: make(map[string]time.Time),
	}
	go p.cleanupLoop()
	return p
}

func (p *InMemoryProvider) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		p.mu.Lock()
		now := time.Now()
		for id, exp := range p.expirationKeys {
			if now.After(exp) {
				delete(p.jobInfos, id)
				delete(p.jobResults, id)
				delete(p.expirationKeys, id)
			}
		}
		p.mu.Unlock()
	}
}

type MemoryJobInfoStore struct{ P *InMemoryProvider }

func (s *MemoryJobInfoStore) Set(ctx context.Context, req domain.ExecutionRequest) error {
	s.P.mu.Lock()
	defer s.P.mu.Unlock()
	s.P.jobInfos[req.ID] = req
	s.P.expirationKeys[req.ID] = time.Now().Add(s.P.ttl)
	return nil
}

func (s *MemoryJobInfoStore) Get(ctx context.Context, id string) (*domain.ExecutionRequest, error) {
	s.P.mu.RLock()
	defer s.P.mu.RUnlock()
	req, ok := s.P.jobInfos[id]
	if !ok {
		return nil, nil
	}
	return &req, nil
}

type MemoryJobStateStore struct{ P *InMemoryProvider }

func (s *MemoryJobStateStore) Set(ctx context.Context, res domain.ExecutionResponse) error {
	s.P.mu.Lock()
	defer s.P.mu.Unlock()
	s.P.jobResults[res.ID] = res
	s.P.expirationKeys[res.ID] = time.Now().Add(s.P.ttl)
	return nil
}

func (s *MemoryJobStateStore) Get(ctx context.Context, id string) (*domain.ExecutionResponse, error) {
	s.P.mu.RLock()
	defer s.P.mu.RUnlock()
	res, ok := s.P.jobResults[id]
	if !ok {
		return nil, nil
	}
	return &res, nil
}

type MemoryJobQueue struct{ P *InMemoryProvider }

func (s *MemoryJobQueue) Push(ctx context.Context, req domain.ExecutionRequest) error {
	select {
	case s.P.jobQueue <- req:
		return nil
	default:
		return fmt.Errorf("queue full")
	}
}

func (s *MemoryJobQueue) Pop(ctx context.Context) (domain.ExecutionRequest, error) {
	select {
	case req := <-s.P.jobQueue:
		return req, nil
	case <-ctx.Done():
		return domain.ExecutionRequest{}, ctx.Err()
	}
}
