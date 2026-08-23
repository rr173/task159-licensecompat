package model

import "testing"

func TestSummarizeCountsMissingLicenseAsBlocker(t *testing.T) {
	findings := []Finding{
		{Kind: FindingBlocker, Code: "component.missing-license", Components: []string{"dep"}},
		{Kind: FindingWarning, Code: "component.unknown-license", Components: []string{"dep"}},
	}
	components := []Component{{ID: "dep", License: ""}}

	summary := Summarize(findings, components)
	if summary.Blockers != 1 {
		t.Fatalf("missing-license must count as a blocker, got blockers=%d", summary.Blockers)
	}
	if summary.Warnings != 1 {
		t.Fatalf("unknown-license must still count as a warning, got warnings=%d", summary.Warnings)
	}
}
