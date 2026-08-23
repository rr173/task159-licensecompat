package model

import "testing"

func TestSubmissionFingerprintIsAliasIndependent(t *testing.T) {
	canonical := Submission{Name: "app", Release: "1", Components: []Component{{ID: "a", Name: "a", Version: "1", License: "MIT"}}}
	aliased := Submission{Name: "app", Release: "1", Components: []Component{{ID: "a", Name: "a", Version: "1", License: "MITLICENSE"}}}
	a, err := SubmissionFingerprint(canonical)
	if err != nil {
		t.Fatal(err)
	}
	b, err := SubmissionFingerprint(aliased)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("fingerprint depends on license alias spelling: %s %s", a, b)
	}
}
