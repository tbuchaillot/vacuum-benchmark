package contractdiff

import (
	"testing"
)

func TestDiffIdentical(t *testing.T) {
	a := Snapshot{
		Version: "upstream",
		Packages: map[string]Package{
			"motor": {
				Functions: []Function{{Name: "Apply", Signature: "func() int"}},
			},
		},
	}
	b := a
	b.Version = "fork"
	got := Diff(a, b)
	if got.Verdict != VerdictDropIn {
		t.Errorf("verdict = %s, want %s", got.Verdict, VerdictDropIn)
	}
	if len(got.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(got.Rows))
	}
	if got.Rows[0].Status != StatusMatch {
		t.Errorf("status = %s, want %s", got.Rows[0].Status, StatusMatch)
	}
}

func TestDiffAddedOnly(t *testing.T) {
	a := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{{Name: "Apply", Signature: "func() int"}},
	}}}
	b := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{
			{Name: "Apply", Signature: "func() int"},
			{Name: "ApplyAsync", Signature: "func() int"},
		},
	}}}
	got := Diff(a, b)
	if got.Verdict != VerdictMinorDrift {
		t.Errorf("verdict = %s, want %s", got.Verdict, VerdictMinorDrift)
	}
}

func TestDiffSignatureChanged(t *testing.T) {
	a := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{{Name: "Apply", Signature: "func() int"}},
	}}}
	b := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{{Name: "Apply", Signature: "func() (int, error)"}},
	}}}
	got := Diff(a, b)
	if got.Verdict != VerdictBreaking {
		t.Errorf("verdict = %s, want %s", got.Verdict, VerdictBreaking)
	}
	found := false
	for _, r := range got.Rows {
		if r.Status == StatusSignatureChanged {
			found = true
		}
	}
	if !found {
		t.Error("expected a signature-changed row")
	}
}

func TestDiffRemoved(t *testing.T) {
	a := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{{Name: "Apply", Signature: "func() int"}},
	}}}
	b := Snapshot{Packages: map[string]Package{"motor": {}}}
	got := Diff(a, b)
	if got.Verdict != VerdictBreaking {
		t.Errorf("verdict = %s, want %s", got.Verdict, VerdictBreaking)
	}
}
