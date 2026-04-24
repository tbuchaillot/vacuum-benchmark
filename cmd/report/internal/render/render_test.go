package render

import (
	"strings"
	"testing"

	"vacuum-benchmark/cmd/report/internal/benchparse"
	"vacuum-benchmark/cmd/report/internal/contractdiff"
)

func TestResultsMarkdownHasHeaderAndOneTablePerRuleset(t *testing.T) {
	upstream := []benchparse.Result{
		{Size: "small", Ruleset: "recommended", NsPerOp: 1_000_000, BytesPerOp: 500_000, AllocsPerOp: 1000},
		{Size: "small", Ruleset: "minimal", NsPerOp: 300_000, BytesPerOp: 100_000, AllocsPerOp: 200},
	}
	fork := []benchparse.Result{
		{Size: "small", Ruleset: "recommended", NsPerOp: 900_000, BytesPerOp: 450_000, AllocsPerOp: 900},
		{Size: "small", Ruleset: "minimal", NsPerOp: 300_000, BytesPerOp: 100_000, AllocsPerOp: 200},
	}

	meta := Meta{GoVersion: "go1.22", CPU: "Apple M2", Date: "2026-04-24", Count: 5, Upstream: "v0.26.1", Fork: "HEAD", UpstreamBuild: "OK", ForkBuild: "OK"}
	got := Results(meta, upstream, fork)

	for _, want := range []string{"## Ruleset: recommended", "## Ruleset: minimal", "-10", "Apple M2", "go1.22"} {
		if !strings.Contains(got, want) {
			t.Errorf("results.md missing %q; got:\n%s", want, got)
		}
	}
}

func TestContractsMarkdownShowsVerdict(t *testing.T) {
	report := contractdiff.Report{
		Rows: []contractdiff.Row{
			{Package: "motor", Kind: "func", Symbol: "Apply", Upstream: "func() int", Fork: "func() int", Status: contractdiff.StatusMatch},
		},
		Verdict: contractdiff.VerdictDropIn,
	}
	got := Contracts(Meta{Upstream: "v0.26.1", Fork: "HEAD"}, report)
	if !strings.Contains(got, "Verdict: **drop-in**") {
		t.Errorf("contracts.md missing verdict; got:\n%s", got)
	}
	if !strings.Contains(got, "| motor | func | Apply |") {
		t.Errorf("contracts.md missing parity row; got:\n%s", got)
	}
}

func TestContractsMarkdownPackageRow(t *testing.T) {
	// Package-level drift (the fork removing an entire package) must render
	// cleanly, not as an empty/malformed row.
	report := contractdiff.Report{
		Rows: []contractdiff.Row{
			{Package: "motor", Kind: "package", Symbol: "motor", Upstream: "present", Status: contractdiff.StatusRemoved},
		},
		Verdict: contractdiff.VerdictBreaking,
	}
	got := Contracts(Meta{}, report)
	if !strings.Contains(got, "| motor | package | motor |") {
		t.Errorf("contracts.md missing package-level row; got:\n%s", got)
	}
}

func TestResultsHandlesBuildFailure(t *testing.T) {
	meta := Meta{UpstreamBuild: "OK", ForkBuild: "BUILD_FAILED: undefined motor.ApplyRulesToRuleSet"}
	got := Results(meta, nil, nil)
	if !strings.Contains(got, "BUILD_FAILED") {
		t.Errorf("expected BUILD_FAILED surface; got:\n%s", got)
	}
}
