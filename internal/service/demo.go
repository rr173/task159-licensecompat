package service

import (
	"context"
	"fmt"
	"licensecompat.local/internal/model"
)

func (s *Service) RunDemo(ctx context.Context) error {
	policy, err := s.CreatePolicy(ctx, model.Policy{Name: "demo delivery policy", Version: 1, Rules: []model.PolicyRule{{ID: "copyleft", Name: "reject strong copyleft", Forbidden: []string{"GPL-3.0"}, AllowWaiver: true}, {ID: "notice", Name: "retain notices", RequireNote: []string{"MIT", "APACHE-2.0"}}}})
	if err != nil {
		return err
	}
	if _, err = s.ActivatePolicy(ctx, policy.ID); err != nil {
		return err
	}
	submission, duplicate, err := s.Submit(ctx, model.Submission{Name: "demo-service", Release: "1.0.0", Components: []model.Component{{ID: "app", Name: "demo service", Version: "1.0.0", License: "PROPRIETARY"}, {ID: "library", Name: "utility", Version: "3.0.0", License: "GPL-3.0"}}, Edges: []model.Edge{{From: "app", To: "library", Scope: "runtime"}}})
	if err != nil {
		return err
	}
	if duplicate {
		return fmt.Errorf("unexpected duplicate demo submission")
	}
	decision, err := s.Analyze(ctx, submission.ID)
	if err != nil {
		return err
	}
	if len(decision.Findings) == 0 {
		return fmt.Errorf("demo produced no findings")
	}
	blockers := make([]string, 0)
	for _, finding := range decision.Findings {
		if finding.Kind == model.FindingBlocker {
			blockers = append(blockers, finding.ID)
		}
	}
	if len(blockers) == 0 {
		return fmt.Errorf("demo expected a blocker")
	}
	for _, blocker := range blockers {
		waiver, err := s.RequestWaiver(ctx, decision.Analysis.ID, blocker, "approved demo exception", nil)
		if err != nil {
			return err
		}
		if _, err = s.ApproveWaiver(ctx, waiver.ID); err != nil {
			return err
		}
	}
	_, err = s.Publish(ctx, decision.Analysis.ID)
	return err
}
