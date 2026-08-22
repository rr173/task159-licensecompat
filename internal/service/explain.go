package service

import (
	"context"
	"fmt"
	"licensecompat.local/internal/analysis"
)

func (s *Service) Explain(ctx context.Context, analysisID string) ([]analysis.Explanation, error) {
	decision, err := s.Decision(ctx, analysisID)
	if err != nil {
		return nil, err
	}
	submission, err := s.store.Submission(ctx, decision.Analysis.SubmissionID)
	if err != nil {
		return nil, err
	}
	out := make([]analysis.Explanation, 0, len(decision.Findings))
	for _, finding := range decision.Findings {
		out = append(out, analysis.Explain(finding, submission))
	}
	return out, nil
}
func (s *Service) Audit(ctx context.Context, subjectType, subjectID string) (any, error) {
	events, err := s.store.Audit(ctx, subjectType, subjectID)
	if err != nil {
		return nil, fmt.Errorf("read audit: %w", err)
	}
	return events, nil
}
