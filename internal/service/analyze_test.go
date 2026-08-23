package service

import (
	"context"
	"licensecompat.local/internal/model"
	"licensecompat.local/internal/store"
	"testing"
	"time"
)

// TestFindingsReturnedInStableAscendingOrder guards the contract that findings
// are exposed in stable, ascending-by-ID order regardless of the order rows
// come back from the database. Findings are inserted in descending ID order to
// simulate an arbitrary/unstable DB row order; both read paths must collapse
// that to a stable ascending sequence across repeated reads.
func TestFindingsReturnedInStableAscendingOrder(t *testing.T) {
	ctx := context.Background()
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	now := time.Now().UTC()

	// Minimal scaffold so SaveAnalysis satisfies foreign-key constraints.
	policy := model.Policy{ID: "pol", Name: "pol", Version: 1, Status: model.PolicyDraft, CreatedAt: now, UpdatedAt: now}
	if err := repository.SavePolicy(ctx, policy); err != nil {
		t.Fatal(err)
	}
	submission := model.Submission{ID: "sub", Name: "app", Release: "1", Fingerprint: "fp", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}, CreatedAt: now}
	if _, _, err := repository.SaveSubmission(ctx, submission); err != nil {
		t.Fatal(err)
	}
	analysis := model.Analysis{ID: "an", SubmissionID: "sub", PolicyID: "pol", Status: model.AnalysisReviewable, InputSnapshot: "{}", PolicySnapshot: "{}", CreatedAt: now, UpdatedAt: now}
	if err := repository.SaveAnalysis(ctx, analysis); err != nil {
		t.Fatal(err)
	}

	// Insert in descending ID order to mimic an arbitrary DB return order.
	out := []model.Finding{
		{ID: "an:code:c", AnalysisID: "an", Kind: model.FindingWarning, Code: "code", Message: "m", CreatedAt: now},
		{ID: "an:code:b", AnalysisID: "an", Kind: model.FindingWarning, Code: "code", Message: "m", CreatedAt: now},
		{ID: "an:code:a", AnalysisID: "an", Kind: model.FindingWarning, Code: "code", Message: "m", CreatedAt: now},
	}
	if err := repository.ReplaceFindings(ctx, "an", out); err != nil {
		t.Fatal(err)
	}

	svc := New(repository)
	want := []string{"an:code:a", "an:code:b", "an:code:c"}

	assertOrder := func(t *testing.T, got []model.Finding) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("got %d findings, want %d (%v)", len(got), len(want), got)
		}
		for i, id := range want {
			if got[i].ID != id {
				t.Fatalf("position %d: got %q, want %q (full: %v)", i, got[i].ID, id, got)
			}
		}
	}

	// Repeated reads must not flip the order or depend on DB row order.
	for i := 0; i < 3; i++ {
		findings, err := svc.Findings(ctx, "an")
		if err != nil {
			t.Fatalf("read %d: %v", i, err)
		}
		assertOrder(t, findings)
	}

	decision, err := svc.Decision(ctx, "an")
	if err != nil {
		t.Fatal(err)
	}
	assertOrder(t, decision.Findings)
}
