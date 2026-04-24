package extract

import (
	"path/filepath"
	"testing"
)

func TestExtractFixture(t *testing.T) {
	fixtureDir, err := filepath.Abs("../../testdata/fixture")
	if err != nil {
		t.Fatal(err)
	}

	got, err := Extract(fixtureDir, []string{"fixture"})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	pkg, ok := got.Packages["fixture"]
	if !ok {
		t.Fatalf("package 'fixture' missing; got keys: %v", keys(got.Packages))
	}

	wantFuncs := map[string]string{
		"ApplyThing": "func(t *Thing) *Thing",
	}
	for name, sig := range wantFuncs {
		found := false
		for _, f := range pkg.Functions {
			if f.Name == name {
				found = true
				if f.Signature != sig {
					t.Errorf("func %s signature = %q, want %q", name, f.Signature, sig)
				}
			}
		}
		if !found {
			t.Errorf("func %s not found; got funcs: %v", name, pkg.Functions)
		}
	}

	for _, f := range pkg.Functions {
		if f.Name == "privateHelper" {
			t.Errorf("unexported privateHelper leaked into output")
		}
	}

	foundThing := false
	for _, typ := range pkg.Types {
		if typ.Name == "Thing" {
			foundThing = true
			if typ.Kind != "struct" {
				t.Errorf("Thing kind = %q, want 'struct'", typ.Kind)
			}
			if len(typ.Fields) != 2 {
				t.Errorf("Thing fields = %d, want 2", len(typ.Fields))
			}
		}
	}
	if !foundThing {
		t.Error("type Thing not found")
	}

	foundDoer := false
	for _, typ := range pkg.Types {
		if typ.Name == "Doer" {
			foundDoer = true
			if typ.Kind != "interface" {
				t.Errorf("Doer kind = %q, want 'interface'", typ.Kind)
			}
		}
	}
	if !foundDoer {
		t.Error("type Doer not found")
	}
}

func keys(m map[string]Package) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
