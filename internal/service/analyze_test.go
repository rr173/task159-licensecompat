package service

import (
	"context"
	"errors"
	"licensecompat.local/internal/model"
	"licensecompat.local/internal/store"
	"testing"
	"time"
)

func seedPublishedAnalysis(t *testing.T) (*Service, string, []model.Finding) {
	t.Helper()
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	svc := New(repository)
	ctx := context.Background()

	policy, err := svc.CreatePolicy(ctx, model.Policy{Name: "publish-freeze", Version: 1, Rules: []model.PolicyRule{{ID: "gpl", Name: "no gpl", Forbidden: []string{"GPL-3.0"}, AllowWaiver: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ActivatePolicy(ctx, policy.ID); err != nil {
		t.Fatal(err)
	}
	submission, _, err := svc.Submit(ctx, model.Submission{
		Name:    "app",
		Release: "1",
		Components: []model.Component{
			{ID: "app", Name: "app", Version: "1", License: "PROPRIETARY"},
			{ID: "dep", Name: "dep", Version: "1", License: "GPL-3.0"},
		},
		Edges: []model.Edge{{From: "app", To: "dep", Scope: "runtime"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := svc.Analyze(ctx, submission.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.Findings) == 0 {
		t.Fatal("expected findings before publish")
	}
	original := append([]model.Finding(nil), decision.Findings...)

	for _, finding := range decision.Findings {
		if finding.Kind == model.FindingBlocker {
			waiver, err := svc.RequestWaiver(ctx, decision.Analysis.ID, finding.ID, "approved exception", nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := svc.ApproveWaiver(ctx, waiver.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := svc.Publish(ctx, decision.Analysis.ID); err != nil {
		t.Fatal(err)
	}
	return svc, decision.Analysis.ID, original
}

func TestReplaceFindingsFrozenAfterPublish(t *testing.T) {
	svc, analysisID, original := seedPublishedAnalysis(t)
	ctx := context.Background()

	replacement := []model.Finding{{
		ID:         "replacement",
		AnalysisID: analysisID,
		Kind:        model.FindingWarning,
		Code:       "post.publish.tamper",
		Message:    "should not persist",
		Components:  []string{"dep"},
		CreatedAt:  time.Now().UTC(),
	}}

	err := svc.ReplaceFindings(ctx, analysisID, replacement)
	if !errors.Is(err, model.ErrImmutable) {
		t.Fatalf("expected ErrImmutable, got %v", err)
	}

	after, err := svc.Findings(ctx, analysisID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(original) {
		t.Fatalf("findings mutated: got %d want %d", len(after), len(original))
	}
	seen := map[string]bool{}
	for _, finding := range original {
		seen[finding.ID] = true
	}
	for _, finding := range after {
		if finding.Code == "post.publish.tamper" {
			t.Fatalf("replacement finding leaked into published analysis: %s", finding.ID)
		}
		if !seen[finding.ID] {
			t.Fatalf("unexpected finding %s after rejected replace", finding.ID)
		}
	}
}

func TestStoreReplaceFindingsRejectsPublished(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	ctx := context.Background()

	policy := model.Policy{ID: "p", Name: "p", Version: 1, Status: model.PolicyActive, Rules: []model.PolicyRule{{ID: "gpl", Name: "no gpl", Forbidden: []string{"GPL-3.0"}, AllowWaiver: true}}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := repository.SavePolicy(ctx, policy); err != nil {
		t.Fatal(err)
	}
	submission := model.Submission{ID: "sub", Name: "app", Release: "1", Fingerprint: "fp", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "PROPRIETARY"}, {ID: "dep", Name: "dep", Version: "1", License: "GPL-3.0"}}, Edges: []model.Edge{{From: "app", To: "dep", Scope: "runtime"}}, CreatedAt: time.Now().UTC()}
	if _, _, err := repository.SaveSubmission(ctx, submission); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	analysis := model.Analysis{ID: "ana", SubmissionID: "sub", PolicyID: "p", Status: model.AnalysisReviewable, InputSnapshot: "{}", PolicySnapshot: "{}", CreatedAt: now, UpdatedAt: now}
	if err := repository.SaveAnalysis(ctx, analysis); err != nil {
		t.Fatal(err)
	}
	original := model.Finding{ID: "ana:original", AnalysisID: "ana", Kind: model.FindingBlocker, Code: "component.original", Message: "original", Components: []string{"dep"}, CreatedAt: now}
	if err := repository.ReplaceFindings(ctx, "ana", []model.Finding{original}); err != nil {
		t.Fatal(err)
	}

	published := analysis
	published.Status = model.AnalysisPublished
	published.UpdatedAt = now
	published.PublishedAt = &now
	if err := repository.UpdateAnalysis(ctx, published); err != nil {
		t.Fatal(err)
	}

	tampered := model.Finding{ID: "ana:tampered", AnalysisID: "ana", Kind: model.FindingWarning, Code: "tampered", Message: "should not persist", Components: []string{"dep"}, CreatedAt: now}
	err = repository.ReplaceFindings(ctx, "ana", []model.Finding{tampered})
	if !errors.Is(err, model.ErrImmutable) {
		t.Fatalf("expected ErrImmutable at store layer, got %v", err)
	}
	after, err := repository.Findings(ctx, "ana")
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 1 || after[0].ID != original.ID {
		t.Fatalf("original findings not preserved at store layer: %#v", after)
	}
}
