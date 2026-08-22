package service

import (
	"context"
	"fmt"
	"github.com/rr173/task159-licensecompat/internal/model"
	"strings"
	"time"
)

func (s *Service) RequestWaiver(ctx context.Context, analysisID, findingID, reason string, expiresAt *time.Time) (model.Waiver, error) {
	if strings.TrimSpace(reason) == "" {
		return model.Waiver{}, model.Invalid("reason", "must not be empty")
	}
	decision, err := s.Decision(ctx, analysisID)
	if err != nil {
		return model.Waiver{}, err
	}
	var found *model.Finding
	for i := range decision.Findings {
		if decision.Findings[i].ID == findingID {
			found = &decision.Findings[i]
			break
		}
	}
	if found == nil {
		return model.Waiver{}, fmt.Errorf("%w: finding", model.ErrNotFound)
	}
	if !model.CanWaiveFinding(found.Kind) {
		return model.Waiver{}, fmt.Errorf("%w: only blocker findings can be waived", model.ErrInvalidState)
	}
	now := s.clock().UTC()
	value := model.Waiver{ID: newID("waiver", now), AnalysisID: analysisID, FindingID: findingID, Reason: reason, Status: model.WaiverRequested, ExpiresAt: expiresAt, CreatedAt: now, UpdatedAt: now}
	if err := s.store.SaveWaiver(ctx, value); err != nil {
		return model.Waiver{}, err
	}
	return value, nil
}
func (s *Service) ApproveWaiver(ctx context.Context, id string) (model.Waiver, error) {
	value, err := s.store.Waiver(ctx, id)
	if err != nil {
		return model.Waiver{}, err
	}
	if !model.CanTransitionWaiver(value.Status, model.WaiverApproved) {
		return model.Waiver{}, fmt.Errorf("%w: waiver", model.ErrInvalidState)
	}
	now := s.clock().UTC()
	if value.ExpiresAt != nil && value.ExpiresAt.Before(now) {
		return model.Waiver{}, fmt.Errorf("%w: waiver expired", model.ErrInvalidState)
	}
	value.Status = model.WaiverApproved
	value.UpdatedAt = now
	if err := s.store.UpdateWaiver(ctx, value); err != nil {
		return model.Waiver{}, err
	}
	return value, nil
}
func (s *Service) Waivers(ctx context.Context, analysisID string) ([]model.Waiver, error) {
	return s.store.Waivers(ctx, analysisID)
}
