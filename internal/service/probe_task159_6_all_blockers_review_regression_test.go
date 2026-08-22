package service_test

import (
	"context"
	"testing"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/service"
	"licensecompat.local/internal/store"
)

func TestBug06_PublishRequiresWaiversForEveryBlocker(t *testing.T) {
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	app := service.New(repository)
	policy, err := app.CreatePolicy(context.Background(), model.Policy{Name: "block-policy", Version: 1, Rules: []model.PolicyRule{{ID: "gpl", Name: "deny gpl", Forbidden: []string{"GPL-3.0"}, AllowWaiver: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.ActivatePolicy(context.Background(), policy.ID); err != nil {
		t.Fatal(err)
	}
	submission, _, err := app.Submit(context.Background(), model.Submission{Name: "blocked-app", Release: "1", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "PROPRIETARY"}, {ID: "gpl", Name: "gpl", Version: "1", License: "GPL-3.0"}, {ID: "revoked", Name: "revoked", Version: "1", License: "MIT", Status: model.ComponentRevoked}}, Edges: []model.Edge{{From: "app", To: "gpl", Scope: "runtime"}}})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := app.Analyze(context.Background(), submission.ID)
	if err != nil {
		t.Fatal(err)
	}
	blockers := make([]model.Finding, 0)
	for _, finding := range decision.Findings {
		if finding.Kind == model.FindingBlocker {
			blockers = append(blockers, finding)
		}
	}
	if len(blockers) < 2 {
		t.Fatalf("expected multiple blockers, got %#v", decision.Findings)
	}
	waiver, err := app.RequestWaiver(context.Background(), decision.Analysis.ID, blockers[0].ID, "one exception", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.ApproveWaiver(context.Background(), waiver.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Publish(context.Background(), decision.Analysis.ID); err == nil {
		t.Fatal("published with an unwaived blocker")
	}
}
