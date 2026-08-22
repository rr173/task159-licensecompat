package service

import (
	"context"
	"fmt"
	"github.com/rr173/task159-licensecompat/internal/analysis"
	"github.com/rr173/task159-licensecompat/internal/model"
)

func (s *Service) Analyze(ctx context.Context, submissionID string) (model.Decision, error) {
	unlock := s.lock("analysis:" + submissionID)
	defer unlock()
	submission, err := s.store.Submission(ctx, submissionID)
	if err != nil {
		return model.Decision{}, err
	}
	active, err := s.store.ActivePolicy(ctx)
	if err != nil {
		return model.Decision{}, err
	}
	now := s.clock().UTC()
	inputSnapshot, err := model.Snapshot(submission)
	if err != nil {
		return model.Decision{}, err
	}
	policySnapshot, err := model.Snapshot(active)
	if err != nil {
		return model.Decision{}, err
	}
	value := model.Analysis{ID: newID("analysis", now), SubmissionID: submission.ID, PolicyID: active.ID, Status: model.AnalysisQueued, InputSnapshot: inputSnapshot, PolicySnapshot: policySnapshot, CreatedAt: now, UpdatedAt: now}
	if err := s.store.SaveAnalysis(ctx, value); err != nil {
		return model.Decision{}, err
	}
	value.Status = model.AnalysisRunning
	value.UpdatedAt = s.clock().UTC()
	if err := s.store.UpdateAnalysis(ctx, value); err != nil {
		return model.Decision{}, err
	}
	result, err := analysis.Evaluate(submission, active, value.ID, s.clock().UTC())
	if err != nil {
		value.Status = model.AnalysisQueued
		value.UpdatedAt = s.clock().UTC()
		_ = s.store.UpdateAnalysis(ctx, value)
		return model.Decision{}, fmt.Errorf("evaluate: %w", err)
	}
	if err := s.ReplaceFindings(ctx, value.ID, result.Findings); err != nil {
		return model.Decision{}, err
	}
	summary := model.Summarize(result.Findings, submission.Components)
	if summary.Blockers > 0 {
		value.Status = model.AnalysisBlocked
	} else {
		value.Status = model.AnalysisReviewable
	}
	value.UpdatedAt = s.clock().UTC()
	if err := s.store.UpdateAnalysis(ctx, value); err != nil {
		return model.Decision{}, err
	}
	return model.Decision{Analysis: value, Findings: result.Findings}, nil
}

func (s *Service) ReplaceFindings(ctx context.Context, analysisID string, findings []model.Finding) error {
	value, err := s.store.Analysis(ctx, analysisID)
	if err != nil {
		return err
	}
	if !model.CanReplaceFindings(value.Status) {
		return fmt.Errorf("%w: findings are frozen", model.ErrImmutable)
	}
	return s.store.ReplaceFindings(ctx, analysisID, findings)
}
func (s *Service) Decision(ctx context.Context, id string) (model.Decision, error) {
	analysisValue, err := s.store.Analysis(ctx, id)
	if err != nil {
		return model.Decision{}, err
	}
	findings, err := s.store.Findings(ctx, id)
	if err != nil {
		return model.Decision{}, err
	}
	return model.Decision{Analysis: analysisValue, Findings: findings}, nil
}
func (s *Service) Findings(ctx context.Context, id string) ([]model.Finding, error) {
	return s.store.Findings(ctx, id)
}
