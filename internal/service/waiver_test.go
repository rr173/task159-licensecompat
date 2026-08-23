package service

import (
	"context"
	"errors"
	"licensecompat.local/internal/model"
	"licensecompat.local/internal/store"
	"testing"
)

// newBlockedAnalysis seeds a fresh store with an active policy and an analyzed
// submission, returning the resulting decision along with helper finding IDs
// grouped by kind so individual waiver tests can target blockers vs. warnings
// vs. obligations.
func newBlockedAnalysis(t *testing.T) (*Service, model.Decision, blockerIDs) {
	t.Helper()
	repository, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { repository.Close() })
	svc := New(repository)
	ctx := context.Background()
	policy, err := svc.CreatePolicy(ctx, model.Policy{
		Name: "test", Version: 1,
		Rules: []model.PolicyRule{
			{ID: "gpl", Name: "no gpl", Forbidden: []string{"GPL-3.0"}, AllowWaiver: true},
			{ID: "notice", Name: "retain notices", RequireNote: []string{"MIT"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.ActivatePolicy(ctx, policy.ID); err != nil {
		t.Fatal(err)
	}
	submission, _, err := svc.Submit(ctx, model.Submission{
		Name: "app", Release: "1",
		Components: []model.Component{
			{ID: "app", Name: "app", Version: "1", License: "PROPRIETARY"},
			{ID: "dep", Name: "dep", Version: "1", License: "GPL-3.0"},
			{ID: "mit", Name: "mit", Version: "1", License: "MIT"},
			{ID: "mystery", Name: "mystery", Version: "1", License: "INTERNAL-TBD"},
		},
		Edges: []model.Edge{{From: "app", To: "dep", Scope: "runtime"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := svc.Analyze(ctx, submission.ID)
	if err != nil {
		t.Fatal(err)
	}
	return svc, decision, groupFindings(decision.Findings)
}

type blockerIDs struct {
	blockers, warnings, obligations []string
}

func groupFindings(findings []model.Finding) blockerIDs {
	out := blockerIDs{}
	for _, f := range findings {
		switch f.Kind {
		case model.FindingBlocker:
			out.blockers = append(out.blockers, f.ID)
		case model.FindingWarning:
			out.warnings = append(out.warnings, f.ID)
		case model.FindingObligation:
			out.obligations = append(out.obligations, f.ID)
		}
	}
	return out
}

func TestRequestWaiverAllowsBlocker(t *testing.T) {
	svc, decision, ids := newBlockedAnalysis(t)
	if len(ids.blockers) == 0 {
		t.Fatal("expected at least one blocker finding")
	}
	waiver, err := svc.RequestWaiver(context.Background(), decision.Analysis.ID, ids.blockers[0], "legitimate exception", nil)
	if err != nil {
		t.Fatalf("blocker should be waivable, got %v", err)
	}
	if waiver.Status != model.WaiverRequested {
		t.Fatalf("status=%v", waiver.Status)
	}
}

func TestRequestWaiverRejectsWarning(t *testing.T) {
	svc, decision, ids := newBlockedAnalysis(t)
	if len(ids.warnings) == 0 {
		t.Fatal("expected at least one warning finding")
	}
	_, err := svc.RequestWaiver(context.Background(), decision.Analysis.ID, ids.warnings[0], "should fail", nil)
	if !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("warning waiver must be an invalid state error, got %v", err)
	}
}

func TestRequestWaiverRejectsObligation(t *testing.T) {
	svc, decision, ids := newBlockedAnalysis(t)
	if len(ids.obligations) == 0 {
		t.Fatal("expected at least one obligation finding")
	}
	_, err := svc.RequestWaiver(context.Background(), decision.Analysis.ID, ids.obligations[0], "should fail", nil)
	if !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("obligation waiver must be an invalid state error, got %v", err)
	}
}
