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

func TestCanonicalLicenses(t *testing.T) {
	got := CanonicalLicenses([]string{" mit ", "MIT", "GPL-3.0"})
	if len(got) != 2 || got[0] != "GPL-3.0" {
		t.Fatalf("unexpected %#v", got)
	}
}
