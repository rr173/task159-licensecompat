package service

import (
	"context"
	"licensecompat.local/internal/model"
	"licensecompat.local/internal/store"
	"testing"
	"time"
)

func TestRecoverRequeuesRunningAnalysis(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now().UTC()

	policy := model.Policy{ID: "policy", Name: "policy", Version: 1, Status: model.PolicyActive, CreatedAt: now, UpdatedAt: now}
	if err := s.SavePolicy(context.Background(), policy); err != nil {
		t.Fatal(err)
	}
	if err := s.ActivatePolicy(context.Background(), policy.ID, policy); err != nil {
		t.Fatal(err)
	}
	submission := model.Submission{ID: "sub", Name: "app", Release: "1", Fingerprint: "fingerprint", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}, CreatedAt: now}
	if _, _, err := s.SaveSubmission(context.Background(), submission); err != nil {
		t.Fatal(err)
	}

	// Simulate a process that crashed mid-analysis: the row was promoted to
	// running but never reached a terminal state.
	running := model.Analysis{ID: "running-1", SubmissionID: "sub", PolicyID: "policy", Status: model.AnalysisQueued, InputSnapshot: "{}", PolicySnapshot: "{}", CreatedAt: now, UpdatedAt: now}
	if err := s.SaveAnalysis(context.Background(), running); err != nil {
		t.Fatal(err)
	}
	running.Status = model.AnalysisRunning
	running.UpdatedAt = now
	if err := s.UpdateAnalysis(context.Background(), running); err != nil {
		t.Fatal(err)
	}

	app := New(s)
	recovered, err := app.Recover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered) != 1 || recovered[0] != "running-1" {
		t.Fatalf("recovered = %v, want [running-1]", recovered)
	}

	after, err := s.Analysis(context.Background(), "running-1")
	if err != nil {
		t.Fatal(err)
	}
	if after.Status != model.AnalysisQueued {
		t.Fatalf("status = %q, want queued", after.Status)
	}
}
