// Package scenarios defines the benchmark matrix (spec size × ruleset).
// It deliberately has no dependency on vacuum — both runners import this
// package, and they pin different versions of vacuum.
package scenarios

type Size string

const (
	SizeSmall  Size = "small"
	SizeMedium Size = "medium"
	SizeLarge  Size = "large"
)

type Ruleset string

const (
	RulesetRecommended Ruleset = "recommended"
	RulesetAll         Ruleset = "all"
	RulesetOWASP       Ruleset = "owasp"
	RulesetMinimal     Ruleset = "minimal"
	RulesetBig         Ruleset = "big"
)

type Scenario struct {
	Size        Size
	Ruleset     Ruleset
	SpecPath    string // repo-relative, e.g. "testdata/openapis/petstore.yaml"
	RulesetPath string // repo-relative, e.g. "testdata/rulesets/recommended.yaml"
}

var specPaths = map[Size]string{
	SizeSmall:  "testdata/openapis/petstore.yaml",
	SizeMedium: "testdata/openapis/digitalocean.yaml",
	SizeLarge:  "testdata/openapis/github.yaml",
}

var rulesetPaths = map[Ruleset]string{
	RulesetRecommended: "testdata/rulesets/recommended.yaml",
	RulesetAll:         "testdata/rulesets/all.yaml",
	RulesetOWASP:       "testdata/rulesets/owasp.yaml",
	RulesetMinimal:     "testdata/rulesets/minimal.yaml",
	RulesetBig:         "testdata/rulesets/big.yaml",
}

// All returns every (Size, Ruleset) combination, ordered (size, ruleset).
func All() []Scenario {
	sizes := []Size{SizeSmall, SizeMedium, SizeLarge}
	rulesets := []Ruleset{RulesetRecommended, RulesetAll, RulesetOWASP, RulesetMinimal, RulesetBig}
	out := make([]Scenario, 0, len(sizes)*len(rulesets))
	for _, s := range sizes {
		for _, r := range rulesets {
			out = append(out, Scenario{
				Size:        s,
				Ruleset:     r,
				SpecPath:    specPaths[s],
				RulesetPath: rulesetPaths[r],
			})
		}
	}
	return out
}
