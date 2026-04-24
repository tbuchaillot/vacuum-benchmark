// Package render produces markdown reports.
package render

import (
	"fmt"
	"sort"
	"strings"

	"vacuum-benchmark/cmd/report/internal/benchparse"
	"vacuum-benchmark/cmd/report/internal/contractdiff"
)

type Meta struct {
	GoVersion     string
	CPU           string
	Date          string
	Count         int
	Upstream      string // e.g. "v0.26.1"
	Fork          string // e.g. "HEAD@<sha>"
	UpstreamBuild string // "OK" or "BUILD_FAILED: <error>"
	ForkBuild     string
}

// Results renders results.md. Passing nil for upstream or fork is
// allowed when that runner failed to build.
func Results(meta Meta, upstream, fork []benchparse.Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Vacuum Benchmark Results\n\n")
	fmt.Fprintf(&b, "- Date: %s\n", meta.Date)
	fmt.Fprintf(&b, "- Go: %s\n", meta.GoVersion)
	fmt.Fprintf(&b, "- CPU: %s\n", meta.CPU)
	fmt.Fprintf(&b, "- count: %d\n", meta.Count)
	fmt.Fprintf(&b, "- upstream: %s (%s)\n", meta.Upstream, meta.UpstreamBuild)
	fmt.Fprintf(&b, "- fork:     %s (%s)\n", meta.Fork, meta.ForkBuild)
	b.WriteString("\n")

	if strings.HasPrefix(meta.UpstreamBuild, "BUILD_FAILED") || strings.HasPrefix(meta.ForkBuild, "BUILD_FAILED") {
		b.WriteString("> One or more runners failed to build. Tables below show whichever data exists.\n\n")
	}

	rulesets := rulesetsPresent(upstream, fork)
	for _, r := range rulesets {
		fmt.Fprintf(&b, "## Ruleset: %s\n\n", r)
		b.WriteString("| Size | upstream ns/op | fork ns/op | Δ | upstream B/op | fork B/op | Δ | upstream allocs | fork allocs | Δ |\n")
		b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n")

		sizes := []string{"small", "medium", "large"}
		for _, s := range sizes {
			u := findResult(upstream, s, r)
			f := findResult(fork, s, r)
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				s,
				formatInt(u.NsPerOp), formatInt(f.NsPerOp), delta(u.NsPerOp, f.NsPerOp),
				formatInt(u.BytesPerOp), formatInt(f.BytesPerOp), delta(u.BytesPerOp, f.BytesPerOp),
				formatInt(u.AllocsPerOp), formatInt(f.AllocsPerOp), delta(u.AllocsPerOp, f.AllocsPerOp),
			)
		}
		b.WriteString("\n")
	}
	b.WriteString("Δ = (fork − upstream) / upstream × 100. Negative = fork is faster/lighter.\n")
	return b.String()
}

// Contracts renders contracts.md.
func Contracts(meta Meta, report contractdiff.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Vacuum Fork Contract Report\n\n")
	fmt.Fprintf(&b, "- upstream: %s\n", meta.Upstream)
	fmt.Fprintf(&b, "- fork:     %s\n\n", meta.Fork)
	fmt.Fprintf(&b, "Verdict: **%s**\n\n", report.Verdict)

	switch report.Verdict {
	case contractdiff.VerdictDropIn:
		b.WriteString("Every analysis-related symbol matches. The fork's public API is identical to upstream's in every package under review; a consumer switching from upstream to fork at this revision should not need to change any caller code.\n\n")
	case contractdiff.VerdictMinorDrift:
		b.WriteString("The fork has added symbols but changed nothing upstream callers already use. Existing callers remain source-compatible; fork-only features are opt-in.\n\n")
	case contractdiff.VerdictBreaking:
		b.WriteString("The fork has renamed, removed, or re-signed symbols that upstream callers use, OR the module path itself has changed. Consumers cannot swap the libraries without updating their import paths and/or call sites. See the parity table for specifics.\n\n")
	}

	b.WriteString("## Parity table\n\n")
	b.WriteString("| Package | Kind | Symbol | Upstream | Fork | Status |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, r := range report.Rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
			r.Package, r.Kind, r.Symbol, escapePipes(r.Upstream), escapePipes(r.Fork), r.Status)
	}
	return b.String()
}

func rulesetsPresent(a, b []benchparse.Result) []string {
	seen := map[string]struct{}{}
	for _, r := range a {
		seen[r.Ruleset] = struct{}{}
	}
	for _, r := range b {
		seen[r.Ruleset] = struct{}{}
	}
	order := []string{"recommended", "all", "owasp", "minimal", "big"}
	var out []string
	for _, r := range order {
		if _, ok := seen[r]; ok {
			out = append(out, r)
		}
	}
	var extras []string
	for r := range seen {
		known := false
		for _, o := range order {
			if r == o {
				known = true
				break
			}
		}
		if !known {
			extras = append(extras, r)
		}
	}
	sort.Strings(extras)
	return append(out, extras...)
}

func findResult(rs []benchparse.Result, size, ruleset string) benchparse.Result {
	for _, r := range rs {
		if r.Size == size && r.Ruleset == ruleset {
			return r
		}
	}
	return benchparse.Result{}
}

func formatInt(n int64) string {
	if n == 0 {
		return "—"
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func delta(upstream, fork int64) string {
	if upstream == 0 || fork == 0 {
		return "—"
	}
	pct := float64(fork-upstream) / float64(upstream) * 100
	return fmt.Sprintf("%+.1f%%", pct)
}

func escapePipes(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}
