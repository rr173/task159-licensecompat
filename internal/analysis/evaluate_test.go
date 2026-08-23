package analysis

import (
	"licensecompat.local/internal/model"
	"testing"
	"time"
)

func TestEvaluateMarksMissingLicenseAsBlocker(t *testing.T) {
	submission := model.Submission{
		Name:    "app",
		Release: "1",
		Components: []model.Component{
			{ID: "app", Name: "app", Version: "1", License: "MIT"},
			{ID: "nolicense", Name: "unlicensed", Version: "1", License: ""},
		},
	}
	policy := model.Policy{ID: "p", Name: "policy", Version: 1}
	now := time.Now().UTC()

	result, err := Evaluate(submission, policy, "analysis-1", now)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	var missing *model.Finding
	for i := range result.Findings {
		if result.Findings[i].Code == "component.missing-license" {
			missing = &result.Findings[i]
			break
		}
	}
	if missing == nil {
		t.Fatalf("expected a component.missing-license finding, got %#v", result.Findings)
	}
	if missing.Kind != model.FindingBlocker {
		t.Fatalf("missing-license must be a blocker, got %s", missing.Kind)
	}
	if missing.Components[0] != "nolicense" {
		t.Fatalf("missing-license finding must reference the unlicensed component, got %#v", missing.Components)
	}

	summary := model.Summarize(result.Findings, submission.Components)
	if summary.Blockers == 0 {
		t.Fatalf("missing-license must count as a blocker so the analysis enters blocked state, got %#v", summary)
	}
	if summary.Warnings != 0 {
		t.Fatalf("missing-license must not be counted as a warning, got %#v", summary)
	}
}
