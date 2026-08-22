package service_test

import (
	"context"
	"testing"
	"time"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug02_QueuedAnalysisCannotBePublished(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	policy, err := app.CreatePolicy(context.Background(), model.Policy{Name: "publish-policy", Version: 1, Rules: []model.PolicyRule{{ID: "notice", Name: "notice", RequireNote: []string{"MIT"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.ActivatePolicy(context.Background(), policy.ID); err != nil {
		t.Fatal(err)
	}
	submission, _, err := app.Submit(context.Background(), model.Submission{Name: "queued-app", Release: "1", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	input, _ := model.Snapshot(submission)
	policySnapshot, _ := model.Snapshot(policy)
	analysis := model.Analysis{ID: "queued-analysis", SubmissionID: submission.ID, PolicyID: policy.ID, Status: model.AnalysisQueued, InputSnapshot: input, PolicySnapshot: policySnapshot, CreatedAt: now, UpdatedAt: now}
	if err := repository.SaveAnalysis(context.Background(), analysis); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Publish(context.Background(), analysis.ID); err == nil {
		t.Fatal("queued analysis was published")
	}
	current, err := repository.Analysis(context.Background(), analysis.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != model.AnalysisQueued {
		t.Fatalf("queued analysis changed to %s", current.Status)
	}
}
