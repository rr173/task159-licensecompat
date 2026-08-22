package service_test

import (
	"context"
	"testing"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug01_SubmissionLicenseAliasesRemainIdempotent(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	first, duplicate, err := app.Submit(context.Background(), model.Submission{Name: "archive", Release: "1", Components: []model.Component{{ID: "dep", Name: "dep", Version: "1", License: "GPL3"}}})
	if err != nil || duplicate {
		t.Fatalf("first submit: %#v duplicate=%v err=%v", first, duplicate, err)
	}
	second, duplicate, err := app.Submit(context.Background(), model.Submission{Name: "archive", Release: "1", Components: []model.Component{{ID: "dep", Name: "dep", Version: "1", License: "GPL-3.0"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !duplicate || second.ID != first.ID {
		t.Fatalf("alias submission was not deduplicated: first=%s second=%s duplicate=%v", first.ID, second.ID, duplicate)
	}
}
