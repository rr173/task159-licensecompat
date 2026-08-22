package service_test

import (
	"context"
	"testing"
	"time"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug07_OnlyBlockersCanReceiveWaivers(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	policy, err := app.CreatePolicy(context.Background(), model.Policy{Name: "waiver-policy", Version: 1, Rules: []model.PolicyRule{{ID: "notice", Name: "notice", RequireNote: []string{"MIT"}}}})
	if err != nil {
		t.Fatal(err)
	}
	submission, _, err := app.Submit(context.Background(), model.Submission{Name: "warning-app", Release: "1", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "CUSTOM"}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	input, _ := model.Snapshot(submission)
	policySnapshot, _ := model.Snapshot(policy)
	item := model.Analysis{ID: "warning-analysis", SubmissionID: submission.ID, PolicyID: policy.ID, Status: model.AnalysisReviewable, InputSnapshot: input, PolicySnapshot: policySnapshot, CreatedAt: now, UpdatedAt: now}
	if err := repository.SaveAnalysis(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	finding := model.Finding{ID: "warning-finding", AnalysisID: item.ID, Kind: model.FindingWarning, Code: "unknown", Message: "needs review", Components: []string{"app"}, CreatedAt: now}
	if err := repository.ReplaceFindings(context.Background(), item.ID, []model.Finding{finding}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.RequestWaiver(context.Background(), item.ID, finding.ID, "not a blocker", nil); err == nil {
		t.Fatal("warning finding accepted a waiver")
	}
}
