package service

import (
	"context"
	"fmt"
	"github.com/rr173/task159-licensecompat/internal/model"
)

type Health struct {
	Status       string `json:"status"`
	Recoverable  int    `json:"recoverable"`
	ActivePolicy bool   `json:"active_policy"`
	Database     string `json:"database"`
}

func (s *Service) SelfCheck(ctx context.Context) (Health, error) {
	if err := s.store.DB().PingContext(ctx); err != nil {
		return Health{}, fmt.Errorf("ping database: %w", err)
	}
	items, err := s.store.RecoverableAnalyses(ctx)
	if err != nil {
		return Health{}, err
	}
	_, activeErr := s.store.ActivePolicy(ctx)
	if activeErr != nil && !model.IsNotFound(activeErr) {
		return Health{}, activeErr
	}
	return Health{Status: "ok", Recoverable: len(items), ActivePolicy: activeErr == nil, Database: "sqlite"}, nil
}
