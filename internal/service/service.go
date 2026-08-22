package service

import (
	"context"
	"github.com/rr173/task159-licensecompat/internal/store"
	"sync"
	"time"
)

type Service struct {
	store *store.Store
	clock func() time.Time
	locks sync.Map
}

func New(repository *store.Store) *Service { return &Service{store: repository, clock: time.Now} }
func (s *Service) Store() *store.Store     { return s.store }
func (s *Service) lock(key string) func() {
	value, _ := s.locks.LoadOrStore(key, &sync.Mutex{})
	mutex := value.(*sync.Mutex)
	mutex.Lock()
	return mutex.Unlock
}
func (s *Service) Recover(ctx context.Context) ([]string, error) {
	items, err := s.store.RecoverableAnalyses(ctx)
	if err != nil {
		return nil, err
	}
	recovered := make([]string, 0, len(items))
	for _, item := range items {
		if item.Status == "running" {
			item.Status = "queued"
			item.UpdatedAt = s.clock().UTC()
			if err := s.store.UpdateAnalysis(ctx, item); err != nil {
				return nil, err
			}
		}
		recovered = append(recovered, item.ID)
	}
	return recovered, nil
}
