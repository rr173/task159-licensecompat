package service_test

import (
	"context"
	"testing"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug08_MissingLicenseBlocksAnalysis(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	policy, err := app.CreatePolicy(context.Background(), model.Policy{Name: "missing-policy", Version: 1, Rules: []model.PolicyRule{{ID: "notice", Name: "notice", RequireNote: []string{"MIT"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.ActivatePolicy(context.Background(), policy.ID); err != nil {
		t.Fatal(err)
	}
	submission, _, err := app.Submit(context.Background(), model.Submission{Name: "missing-license-app", Release: "1", Components: []model.Component{{ID: "app", Name: "app", Version: "1"}}})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := app.Analyze(context.Background(), submission.ID)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Analysis.Status != model.AnalysisBlocked {
		t.Fatalf("missing license status=%s findings=%#v", decision.Analysis.Status, decision.Findings)
	}
}
