package service_test

import (
	"context"
	"testing"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug05_CompareRequiresBothAnalysisIDs(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	if _, err := app.Compare(context.Background(), "", "right"); err == nil || !model.IsInvalid(err) {
		t.Fatalf("expected invalid comparison identifiers, got %v", err)
	}
}
