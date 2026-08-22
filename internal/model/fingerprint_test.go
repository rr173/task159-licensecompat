package model

import "testing"

func TestSubmissionFingerprintIsOrderIndependent(t *testing.T) {
	left := Submission{Name: "app", Release: "1", Components: []Component{{ID: "b", Name: "b", Version: "1", License: "MIT"}, {ID: "a", Name: "a", Version: "1", License: "GPL-3.0"}}, Edges: []Edge{{From: "b", To: "a", Scope: "runtime"}}}
	right := Submission{Name: "app", Release: "1", Components: []Component{{ID: "a", Name: "a", Version: "1", License: "GPL-3.0"}, {ID: "b", Name: "b", Version: "1", License: "MIT"}}, Edges: []Edge{{From: "b", To: "a", Scope: "runtime"}}}
	a, err := SubmissionFingerprint(left)
	if err != nil {
		t.Fatal(err)
	}
	b, err := SubmissionFingerprint(right)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("fingerprint depends on order: %s %s", a, b)
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	value, err := Snapshot(Policy{Name: "p", Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	restored, err := Restore[Policy](value)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Name != "p" || restored.Version != 1 {
		t.Fatalf("unexpected %#v", restored)
	}
}

func TestCanonicalLicenseAliases(t *testing.T) {
	if CanonicalLicense("GPL3") != "GPL-3.0" || CanonicalLicense("apache 2.0") != "APACHE-2.0" {
		t.Fatal("license aliases must have one canonical spelling")
	}
}
