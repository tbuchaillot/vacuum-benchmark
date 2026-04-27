package bench_test

import (
	"testing"

	"github.com/daveshanley/vacuum/motor"

	"vacuum-benchmark/internal/scenarios"
)

func BenchmarkLint(b *testing.B) {
	for _, sc := range scenarios.All() {
		sc := sc
		b.Run(string(sc.Size)+"/"+string(sc.Ruleset), func(b *testing.B) {
			specBytes := mustReadFile(b, sc.SpecPath)
			rs := buildRuleset(b, sc.RulesetPath)

			b.ReportAllocs()
			b.SetBytes(int64(len(specBytes)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				res := motor.ApplyRulesToRuleSet(&motor.RuleSetExecution{
					RuleSet: rs,
					Spec:    specBytes,
				})
				if res == nil {
					b.Fatal("nil result from ApplyRulesToRuleSet")
				}
			}
		})
	}
}
