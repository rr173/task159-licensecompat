package service_test

import (
	"context"
	"testing"
	"time"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug09_FindingsHaveStableAscendingOrder(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	policy, err := app.CreatePolicy(context.Background(), model.Policy{Name: "order-policy", Version: 1, Rules: []model.PolicyRule{{ID: "notice", Name: "notice", RequireNote: []string{"MIT"}}}})
	if err != nil {
		t.Fatal(err)
	}
	submission, _, err := app.Submit(context.Background(), model.Submission{Name: "order-app", Release: "1", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	input, _ := model.Snapshot(submission)
	policySnapshot, _ := model.Snapshot(policy)
	item := model.Analysis{ID: "ordered-analysis", SubmissionID: submission.ID, PolicyID: policy.ID, Status: model.AnalysisReviewable, InputSnapshot: input, PolicySnapshot: policySnapshot, CreatedAt: now, UpdatedAt: now}
	if err := repository.SaveAnalysis(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	findings := []model.Finding{{ID: "finding-a", AnalysisID: item.ID, Kind: model.FindingWarning, Code: "a", Message: "a", CreatedAt: now}, {ID: "finding-b", AnalysisID: item.ID, Kind: model.FindingWarning, Code: "b", Message: "b", CreatedAt: now}}
	if err := repository.ReplaceFindings(context.Background(), item.ID, findings); err != nil {
		t.Fatal(err)
	}
	decision, err := app.Decision(context.Background(), item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.Findings) != 2 || decision.Findings[0].ID != "finding-a" || decision.Findings[1].ID != "finding-b" {
		t.Fatalf("unstable findings order: %#v", decision.Findings)
	}
}
