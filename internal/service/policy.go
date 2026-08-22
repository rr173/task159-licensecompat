package service

import (
	"context"
	"fmt"
	"github.com/rr173/task159-licensecompat/internal/model"
)

func (s *Service) CreatePolicy(ctx context.Context, value model.Policy) (model.Policy, error) {
	now := s.clock().UTC()
	if value.ID == "" {
		value.ID = newID("pol", now)
	}
	if value.Status == "" {
		value.Status = model.PolicyDraft
	}
	value.CreatedAt = now
	value.UpdatedAt = now
	if err := model.ValidatePolicy(value); err != nil {
		return model.Policy{}, err
	}
	if err := s.store.SavePolicy(ctx, value); err != nil {
		return model.Policy{}, fmt.Errorf("save policy: %w", err)
	}
	return value, nil
}
func (s *Service) Policies(ctx context.Context) ([]model.Policy, error) { return s.store.Policies(ctx) }
func (s *Service) ActivatePolicy(ctx context.Context, id string) (model.Policy, error) {
	unlock := s.lock("policy-active")
	defer unlock()
	value, err := s.store.Policy(ctx, id)
	if err != nil {
		return model.Policy{}, err
	}
	if value.Status == model.PolicyRetired {
		return model.Policy{}, fmt.Errorf("%w: retired policy cannot be activated", model.ErrInvalidState)
	}
	value.Status = model.PolicyActive
	value.UpdatedAt = s.clock().UTC()
	if err := s.store.ActivatePolicy(ctx, id, value); err != nil {
		return model.Policy{}, err
	}
	return value, nil
}
