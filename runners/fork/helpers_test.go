package bench_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buraksekili/vacuum/rulesets"
)

// repoRoot walks up from the current directory until it finds go.work.
// It caches the result on first call. Callers pass *testing.TB so either
// *testing.T or *testing.B can use it.
var cachedRepoRoot string

func repoRoot(tb testing.TB) string {
	tb.Helper()
	if cachedRepoRoot != "" {
		return cachedRepoRoot
	}
	dir, err := os.Getwd()
	if err != nil {
		tb.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			cachedRepoRoot = dir
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			tb.Fatalf("go.work not found walking up from %s", dir)
		}
		dir = parent
	}
}

func mustReadFile(tb testing.TB, relPath string) []byte {
	tb.Helper()
	abs := filepath.Join(repoRoot(tb), relPath)
	b, err := os.ReadFile(abs)
	if err != nil {
		tb.Fatalf("read %s: %v", abs, err)
	}
	return b
}

func buildRuleset(tb testing.TB, relPath string) *rulesets.RuleSet {
	tb.Helper()
	raw := mustReadFile(tb, relPath)
	store := rulesets.BuildDefaultRuleSets()
	rs, err := rulesets.CreateRuleSetFromData(raw)
	if err != nil {
		tb.Fatalf("parse ruleset %s: %v", relPath, err)
	}
	// Expand any `extends:` into the full ruleset.
	return store.GenerateRuleSetFromSuppliedRuleSet(rs)
}

func TestRepoRootFindsGoWork(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "go.work")); err != nil {
		t.Fatalf("go.work missing at resolved root %s: %v", root, err)
	}
}

func TestMustReadFileReadsPetstore(t *testing.T) {
	// Skips cleanly if testdata not fetched yet; not a helper failure.
	petstore := filepath.Join(repoRoot(t), "testdata/openapis/petstore.yaml")
	if _, err := os.Stat(petstore); err != nil {
		t.Skip("testdata/openapis/petstore.yaml not fetched — run `make fetch`")
	}
	got := mustReadFile(t, "testdata/openapis/petstore.yaml")
	if len(got) == 0 {
		t.Fatal("petstore.yaml empty")
	}
}

func TestBuildRulesetParsesRecommended(t *testing.T) {
	rs := buildRuleset(t, "testdata/rulesets/recommended.yaml")
	if rs == nil {
		t.Fatal("buildRuleset returned nil")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("recommended ruleset resolved to zero rules — extends likely broken")
	}
}
