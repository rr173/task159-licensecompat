package store

import (
	"context"
	"licensecompat.local/internal/model"
	"testing"
	"time"
)

func TestRecoverableAnalysesIncludesRunning(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now().UTC()

	// Seed the foreign-key targets a real analysis depends on.
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

	analysis := func(id string, status model.AnalysisStatus) model.Analysis {
		return model.Analysis{ID: id, SubmissionID: "sub", PolicyID: "policy", Status: status, InputSnapshot: "{}", PolicySnapshot: "{}", CreatedAt: now, UpdatedAt: now}
	}
	for _, a := range []model.Analysis{
		analysis("queued-1", model.AnalysisQueued),
		analysis("running-1", model.AnalysisRunning),
		analysis("blocked-1", model.AnalysisBlocked),
		analysis("reviewable-1", model.AnalysisReviewable),
		analysis("published-1", model.AnalysisPublished),
	} {
		if err := s.SaveAnalysis(context.Background(), a); err != nil {
			t.Fatal(err)
		}
		// SaveAnalysis inserts with the initial status; bump the rest so the
		// terminal-state rows reach the status requested above.
		if a.Status != model.AnalysisQueued {
			if err := s.UpdateAnalysis(context.Background(), a); err != nil {
				t.Fatal(err)
			}
		}
	}

	got, err := s.RecoverableAnalyses(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var ids []string
	for _, item := range got {
		ids = append(ids, item.ID)
	}
	want := []string{"queued-1", "running-1"}
	if len(ids) != len(want) {
		t.Fatalf("recoverable = %v, want %v", ids, want)
	}
	seen := map[string]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	for _, id := range want {
		if !seen[id] {
			t.Fatalf("recoverable = %v, want %v (missing %s)", ids, want, id)
		}
	}
}
