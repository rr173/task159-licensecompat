package service_test

import (
	"context"
	"testing"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug10_PublishedFindingsRemainImmutable(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	policy, err := app.CreatePolicy(context.Background(), model.Policy{Name: "freeze-policy", Version: 1, Rules: []model.PolicyRule{{ID: "notice", Name: "notice", RequireNote: []string{"MIT"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.ActivatePolicy(context.Background(), policy.ID); err != nil {
		t.Fatal(err)
	}
	submission, _, err := app.Submit(context.Background(), model.Submission{Name: "freeze-app", Release: "1", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := app.Analyze(context.Background(), submission.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Publish(context.Background(), decision.Analysis.ID); err != nil {
		t.Fatal(err)
	}
	replacement := model.Finding{ID: "replacement", AnalysisID: decision.Analysis.ID, Kind: model.FindingWarning, Code: "replacement", Message: "must not replace", Components: []string{"app"}}
	if err := app.ReplaceFindings(context.Background(), decision.Analysis.ID, []model.Finding{replacement}); err == nil {
		t.Fatal("published findings were replaced")
	}
	findings, err := app.Findings(context.Background(), decision.Analysis.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range findings {
		if finding.ID == replacement.ID {
			t.Fatal("replacement finding was persisted")
		}
	}
}
