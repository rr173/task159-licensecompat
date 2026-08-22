package policy

import (
	"github.com/rr173/task159-licensecompat/internal/model"
	"testing"
)

func TestEvaluatePolicy(t *testing.T) {
	result := Evaluate(model.Policy{Rules: []model.PolicyRule{{ID: "one", Name: "deny", Forbidden: []string{"GPL-3.0"}, AllowWaiver: true}, {ID: "two", Name: "notice", RequireNote: []string{"MIT"}}}}, []string{"MIT", "GPL3"})
	if len(result) != 2 || result[0].Kind != model.FindingBlocker {
		t.Fatalf("unexpected %#v", result)
	}
}
func TestPairCompatibility(t *testing.T) {
	kind, _, waiver := PairCompatibility("PROPRIETARY", "GPL-3.0")
	if kind != model.FindingBlocker || !waiver {
		t.Fatalf("unexpected result %s %v", kind, waiver)
	}
}
func TestNormalize(t *testing.T) {
	if Normalize(" apache 2.0 ") != "APACHE-2.0" {
		t.Fatal("alias was not normalized")
	}
}
