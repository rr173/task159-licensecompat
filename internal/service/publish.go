package service

import (
	"context"
	"fmt"
	"github.com/rr173/task159-licensecompat/internal/analysis"
	"github.com/rr173/task159-licensecompat/internal/model"
)

func (s *Service) Publish(ctx context.Context, analysisID string) (model.Analysis, error) {
	unlock := s.lock("publish:" + analysisID)
	defer unlock()
	value, err := s.store.Analysis(ctx, analysisID)
	if err != nil {
		return model.Analysis{}, err
	}
	if value.Status == model.AnalysisPublished {
		return model.Analysis{}, fmt.Errorf("%w: analysis already published", model.ErrImmutable)
	}
	if value.Status != model.AnalysisReviewable && value.Status != model.AnalysisBlocked {
		return model.Analysis{}, fmt.Errorf("%w: analysis must be reviewed before publication", model.ErrInvalidState)
	}
	findings, err := s.store.Findings(ctx, analysisID)
	if err != nil {
		return model.Analysis{}, err
	}
	waivers, err := s.store.Waivers(ctx, analysisID)
	if err != nil {
		return model.Analysis{}, err
	}
	approved := map[string]bool{}
	for _, waiver := range waivers {
		if waiver.Status == model.WaiverApproved {
			approved[waiver.FindingID] = true
		}
	}
	for _, finding := range findings {
		if finding.Kind == model.FindingBlocker && !approved[finding.ID] {
			return model.Analysis{}, fmt.Errorf("%w: blocker %s requires an approved waiver", model.ErrInvalidState, finding.ID)
		}
	}
	now := s.clock().UTC()
	value.Status = model.AnalysisPublished
	value.PublishedAt = &now
	value.UpdatedAt = now
	if err := s.store.UpdateAnalysis(ctx, value); err != nil {
		return model.Analysis{}, err
	}
	return value, nil
}
func (s *Service) Compare(ctx context.Context, leftID, rightID string) (any, error) {
	if err := model.ValidateComparisonIDs(leftID, rightID); err != nil {
		return nil, err
	}
	left, err := s.store.Findings(ctx, leftID)
	if err != nil {
		return nil, err
	}
	right, err := s.store.Findings(ctx, rightID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"left": leftID, "right": rightID, "comparison": analysis.Compare(left, right)}, nil
}
func compareFindings(left, right []model.Finding) map[string]any {
	leftSet := map[string]model.Finding{}
	rightSet := map[string]model.Finding{}
	for _, finding := range left {
		leftSet[finding.Code+finding.Message] = finding
	}
	for _, finding := range right {
		rightSet[finding.Code+finding.Message] = finding
	}
	added := []model.Finding{}
	removed := []model.Finding{}
	for key, finding := range rightSet {
		if _, ok := leftSet[key]; !ok {
			added = append(added, finding)
		}
	}
	for key, finding := range leftSet {
		if _, ok := rightSet[key]; !ok {
			removed = append(removed, finding)
		}
	}
	return map[string]any{"added": added, "removed": removed, "changed": len(added) > 0 || len(removed) > 0}
}
