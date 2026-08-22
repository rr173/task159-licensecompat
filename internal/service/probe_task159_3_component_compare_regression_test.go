package service_test

import (
	"context"
	"testing"
	"time"

	"licensecompat.local/internal/analysis"
	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug03_CompareKeepsComponentSpecificFindings(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	policy, err := app.CreatePolicy(context.Background(), model.Policy{Name: "compare-policy", Version: 1, Rules: []model.PolicyRule{{ID: "notice", Name: "notice", RequireNote: []string{"MIT"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.ActivatePolicy(context.Background(), policy.ID); err != nil {
		t.Fatal(err)
	}
	submission, _, err := app.Submit(context.Background(), model.Submission{Name: "compare-app", Release: "1", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	input, _ := model.Snapshot(submission)
	policySnapshot, _ := model.Snapshot(policy)
	left := model.Analysis{ID: "left", SubmissionID: submission.ID, PolicyID: policy.ID, Status: model.AnalysisReviewable, InputSnapshot: input, PolicySnapshot: policySnapshot, CreatedAt: now, UpdatedAt: now}
	right := left
	right.ID = "right"
	if err := repository.SaveAnalysis(context.Background(), left); err != nil {
		t.Fatal(err)
	}
	if err := repository.SaveAnalysis(context.Background(), right); err != nil {
		t.Fatal(err)
	}
	findingLeft := model.Finding{ID: "left-finding", AnalysisID: "left", Kind: model.FindingWarning, Code: "same", Message: "review", Components: []string{"component-a"}, CreatedAt: now}
	findingRight := model.Finding{ID: "right-finding", AnalysisID: "right", Kind: model.FindingWarning, Code: "same", Message: "review", Components: []string{"component-b"}, CreatedAt: now}
	if err := repository.ReplaceFindings(context.Background(), "left", []model.Finding{findingLeft}); err != nil {
		t.Fatal(err)
	}
	if err := repository.ReplaceFindings(context.Background(), "right", []model.Finding{findingRight}); err != nil {
		t.Fatal(err)
	}
	value, err := app.Compare(context.Background(), "left", "right")
	if err != nil {
		t.Fatal(err)
	}
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("unexpected comparison wrapper %T", value)
	}
	comparison, ok := result["comparison"].(analysis.Comparison)
	if !ok {
		t.Fatalf("comparison lost typed component identity: %T", result["comparison"])
	}
	if len(comparison.Added) != 1 || len(comparison.Removed) != 1 || !comparison.Changed {
		t.Fatalf("unexpected comparison %#v", comparison)
	}
}
