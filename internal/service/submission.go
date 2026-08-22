package service

import (
	"context"
	"fmt"
	"licensecompat.local/internal/model"
	"strings"
)

func (s *Service) Submit(ctx context.Context, value model.Submission) (model.Submission, bool, error) {
	now := s.clock().UTC()
	if value.ID == "" {
		value.ID = newID("sub", now)
	}
	value.CreatedAt = now
	for i := range value.Components {
		value.Components[i].License = strings.TrimSpace(value.Components[i].License)
		if value.Components[i].Status == "" {
			value.Components[i].Status = model.ComponentPending
		}
	}
	if err := model.ValidateSubmission(value); err != nil {
		return model.Submission{}, false, err
	}
	fingerprint, err := model.SubmissionFingerprint(value)
	if err != nil {
		return model.Submission{}, false, err
	}
	value.Fingerprint = fingerprint
	stored, duplicate, err := s.store.SaveSubmission(ctx, value)
	if err != nil {
		return model.Submission{}, false, fmt.Errorf("save submission: %w", err)
	}
	return stored, duplicate, nil
}
func (s *Service) Submission(ctx context.Context, id string) (model.Submission, error) {
	return s.store.Submission(ctx, id)
}
func (s *Service) Submissions(ctx context.Context) ([]model.Submission, error) {
	return s.store.Submissions(ctx)
}
