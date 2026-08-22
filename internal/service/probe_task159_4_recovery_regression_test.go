package service_test

import (
	"context"
	"testing"
	"time"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug04_RunningAnalysesRecoverAfterRestart(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	policy, err := app.CreatePolicy(context.Background(), model.Policy{Name: "recovery-policy", Version: 1, Rules: []model.PolicyRule{{ID: "notice", Name: "notice", RequireNote: []string{"MIT"}}}})
	if err != nil {
		t.Fatal(err)
	}
	submission, _, err := app.Submit(context.Background(), model.Submission{Name: "recover-app", Release: "1", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	input, _ := model.Snapshot(submission)
	policySnapshot, _ := model.Snapshot(policy)
	item := model.Analysis{ID: "running-analysis", SubmissionID: submission.ID, PolicyID: policy.ID, Status: model.AnalysisRunning, InputSnapshot: input, PolicySnapshot: policySnapshot, CreatedAt: now, UpdatedAt: now}
	if err := repository.SaveAnalysis(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	recovered, err := app.Recover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered) != 1 || recovered[0] != item.ID {
		t.Fatalf("recovery list=%v", recovered)
	}
	current, err := repository.Analysis(context.Background(), item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != model.AnalysisQueued {
		t.Fatalf("status after recovery=%s", current.Status)
	}
}
