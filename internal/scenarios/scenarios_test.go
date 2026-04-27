package scenarios

import (
	"testing"
)

func TestAllReturns15Scenarios(t *testing.T) {
	got := All()
	if len(got) != 15 {
		t.Fatalf("len(All()) = %d, want 15 (3 sizes x 5 rulesets)", len(got))
	}
}

func TestAllCoversCartesianProduct(t *testing.T) {
	got := All()
	seen := make(map[string]bool)
	for _, sc := range got {
		key := string(sc.Size) + "/" + string(sc.Ruleset)
		if seen[key] {
			t.Errorf("duplicate scenario: %s", key)
		}
		seen[key] = true
	}
	sizes := []Size{SizeSmall, SizeMedium, SizeLarge}
	rulesets := []Ruleset{RulesetRecommended, RulesetAll, RulesetOWASP, RulesetMinimal, RulesetBig}
	for _, s := range sizes {
		for _, r := range rulesets {
			key := string(s) + "/" + string(r)
			if !seen[key] {
				t.Errorf("missing scenario: %s", key)
			}
		}
	}
}

func TestAllUsesRepoRelativePaths(t *testing.T) {
	for _, sc := range All() {
		if sc.SpecPath == "" || sc.RulesetPath == "" {
			t.Errorf("scenario %s/%s has empty paths", sc.Size, sc.Ruleset)
		}
		if sc.SpecPath[0] == '/' || sc.RulesetPath[0] == '/' {
			t.Errorf("scenario %s/%s has absolute path (want repo-relative)", sc.Size, sc.Ruleset)
		}
	}
}
