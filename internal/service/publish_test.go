package service

import (
	"context"
	"errors"
	"licensecompat.local/internal/model"
	"licensecompat.local/internal/store"
	"testing"
	"time"
)

// newQueuedAnalysis seeds a fresh store with an analysis that is still queued
// (i.e. it has not been evaluated/reviewed yet) so Publish can be exercised
// against the pre-review state.
func newQueuedAnalysis(t *testing.T) (*Service, string) {
	t.Helper()
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { repository.Close() })
	now := time.Now().UTC()
	ctx := context.Background()
	policy := model.Policy{ID: "policy-q", Name: "p", Version: 1, Status: model.PolicyActive, CreatedAt: now, UpdatedAt: now}
	if err := repository.SavePolicy(ctx, policy); err != nil {
		t.Fatal(err)
	}
	submission := model.Submission{ID: "sub-q", Name: "app", Release: "1", Fingerprint: "fp", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}, CreatedAt: now}
	if _, _, err := repository.SaveSubmission(ctx, submission); err != nil {
		t.Fatal(err)
	}
	value := model.Analysis{
		ID:           "analysis-queued",
		SubmissionID: submission.ID,
		PolicyID:     policy.ID,
		Status:       model.AnalysisQueued,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repository.SaveAnalysis(ctx, value); err != nil {
		t.Fatal(err)
	}
	return New(repository), value.ID
}

func TestPublishRejectsQueuedAnalysis(t *testing.T) {
	svc, id := newQueuedAnalysis(t)
	_, err := svc.Publish(context.Background(), id)
	if !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("expected ErrInvalidState for queued publish, got %v", err)
	}
	after, err := svc.Store().Analysis(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if after.Status != model.AnalysisQueued {
		t.Fatalf("queued analysis status must be preserved, got %q", after.Status)
	}
	if after.PublishedAt != nil {
		t.Fatalf("queued analysis must not gain a published_at timestamp, got %v", after.PublishedAt)
	}
}
