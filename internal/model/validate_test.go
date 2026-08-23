package model

import "testing"

func TestValidateSubmissionRejectsDuplicateComponents(t *testing.T) {
	err := ValidateSubmission(Submission{Name: "x", Release: "1", Components: []Component{{ID: "a", Name: "a", Version: "1"}, {ID: "a", Name: "b", Version: "1"}}})
	if !IsInvalid(err) {
		t.Fatalf("expected invalid error, got %v", err)
	}
}

func TestAnalysisTransitions(t *testing.T) {
	if !CanTransitionAnalysis(AnalysisQueued, AnalysisRunning) {
		t.Fatal("queued should start")
	}
	if CanTransitionAnalysis(AnalysisPublished, AnalysisQueued) {
		t.Fatal("published must be immutable")
	}
}

func TestCanReplaceFindingsRespectsPublication(t *testing.T) {
	if !CanReplaceFindings(AnalysisQueued) || !CanReplaceFindings(AnalysisBlocked) || !CanReplaceFindings(AnalysisReviewable) {
		t.Fatal("pre-publication analysis must allow findings replacement")
	}
	if CanReplaceFindings(AnalysisPublished) {
		t.Fatal("published analysis must freeze findings")
	}
}

func TestCanonicalLicenses(t *testing.T) {
	got := CanonicalLicenses([]string{" mit ", "MIT", "GPL-3.0"})
	if len(got) != 2 || got[0] != "GPL-3.0" {
		t.Fatalf("unexpected %#v", got)
	}
}
