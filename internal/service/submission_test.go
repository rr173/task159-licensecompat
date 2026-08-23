package service

import (
	"context"
	"licensecompat.local/internal/model"
	"licensecompat.local/internal/store"
	"testing"
	"time"
)

// fixedClock returns a stable timestamp so both submissions share CreatedAt,
// isolating the test from clock drift between the two Submit calls.
type fixedClock struct{ t time.Time }

func (f fixedClock) now() time.Time { return f.t }

func newServiceWithFixedClock(t *testing.T) (*Service, *store.Store, func() time.Time) {
	t.Helper()
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	svc := New(repository)
	clock := fixedClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	svc.clock = clock.now
	return svc, repository, clock.now
}

func TestSubmitNormalizesLicenseAliases(t *testing.T) {
	svc, _, _ := newServiceWithFixedClock(t)
	first := model.Submission{Name: "app", Release: "1", Components: []model.Component{{ID: "a", Name: "a", Version: "1", License: "MITLICENSE"}}}
	stored, duplicate, err := svc.Submit(context.Background(), first)
	if err != nil || duplicate || stored.Components[0].License != "MIT" {
		t.Fatalf("first submit: stored=%#v duplicate=%v err=%v", stored, duplicate, err)
	}
}

func TestSubmitAliasIsIdempotentAndDeduplicates(t *testing.T) {
	svc, _, _ := newServiceWithFixedClock(t)
	canonical := model.Submission{Name: "app", Release: "1", Components: []model.Component{{ID: "a", Name: "a", Version: "1", License: "MIT"}}}
	stored, duplicate, err := svc.Submit(context.Background(), canonical)
	if err != nil || duplicate {
		t.Fatalf("first submit (MIT): stored=%#v duplicate=%v err=%v", stored, duplicate, err)
	}
	originalID := stored.ID

	aliased := model.Submission{Name: "app", Release: "1", Components: []model.Component{{ID: "a", Name: "a", Version: "1", License: "MITLICENSE"}}}
	replay, duplicate, err := svc.Submit(context.Background(), aliased)
	if err != nil || !duplicate {
		t.Fatalf("alias submit must be a duplicate: replay=%#v duplicate=%v err=%v", replay, duplicate, err)
	}
	if replay.ID != originalID {
		t.Fatalf("alias submit returned a different submission id: want %s got %s", originalID, replay.ID)
	}
	if replay.Components[0].License != "MIT" {
		t.Fatalf("duplicate submission license not canonical: %s", replay.Components[0].License)
	}

	// No second row should have been persisted: list must contain exactly one.
	all, err := svc.Submissions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("expected exactly one persisted submission, got %d", len(all))
	}
}

func TestSubmitDifferentAliasesAcrossComponentsDeduplicates(t *testing.T) {
	svc, _, _ := newServiceWithFixedClock(t)
	first := model.Submission{Name: "app", Release: "1", Components: []model.Component{{ID: "a", Name: "a", Version: "1", License: "GPL3"}}}
	stored, _, err := svc.Submit(context.Background(), first)
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}
	second := model.Submission{Name: "app", Release: "1", Components: []model.Component{{ID: "a", Name: "a", Version: "1", License: "GPLV3"}}}
	replay, duplicate, err := svc.Submit(context.Background(), second)
	if err != nil || !duplicate || replay.ID != stored.ID {
		t.Fatalf("GPL3 vs GPLV3 must dedup: replay=%#v duplicate=%v err=%v", replay, duplicate, err)
	}
}
