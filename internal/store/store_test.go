package store

import (
	"context"
	"github.com/rr173/task159-licensecompat/internal/model"
	"testing"
	"time"
)

func TestSubmissionPersistsAndDeduplicates(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now().UTC()
	value := model.Submission{ID: "sub", Name: "app", Release: "1", Fingerprint: "fingerprint", Components: []model.Component{{ID: "app", Name: "app", Version: "1", License: "MIT"}}, CreatedAt: now}
	stored, duplicate, err := s.SaveSubmission(context.Background(), value)
	if err != nil || duplicate || stored.ID != "sub" {
		t.Fatalf("first: %#v %v %v", stored, duplicate, err)
	}
	stored, duplicate, err = s.SaveSubmission(context.Background(), value)
	if err != nil || !duplicate || stored.ID != "sub" {
		t.Fatalf("second: %#v %v %v", stored, duplicate, err)
	}
}
func TestPolicyActivationIsDurable(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now().UTC()
	p := model.Policy{ID: "p", Name: "policy", Version: 1, Status: model.PolicyDraft, CreatedAt: now, UpdatedAt: now}
	if err := s.SavePolicy(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	p.Status = model.PolicyActive
	if err := s.ActivatePolicy(context.Background(), p.ID, p); err != nil {
		t.Fatal(err)
	}
	active, err := s.ActivePolicy(context.Background())
	if err != nil || active.ID != "p" {
		t.Fatalf("%#v %v", active, err)
	}
}
