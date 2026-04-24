# vacuum-benchmark

A reproducible Go benchmark comparing two OpenAPI linter implementations:

- **Upstream:** [`github.com/daveshanley/vacuum`](https://github.com/daveshanley/vacuum) pinned to `v0.26.1`
- **Fork:** [`github.com/buraksekili/vacuum`](https://github.com/buraksekili/vacuum) pinned to commit `b857c8f` (2025-07)

Measures lint performance across spec size × ruleset combinations and diffs the analysis-related public API surface.

See `docs/superpowers/specs/2026-04-24-vacuum-benchmark-design.md` for the full design.

## Quick Start

```bash
make fetch      # downloads pinned OpenAPI specs into testdata/openapis/
make bench      # runs benchmarks + contract diff; writes results/results.md + results/contracts.md
```

For more statistical stability, raise the iteration count:

```bash
make bench COUNT=10
```

## What gets measured

### Performance (`results/results.md`)

For each of 3 spec sizes × 5 rulesets = 15 scenarios, `BenchmarkLint` measures end-to-end `motor.ApplyRulesToRuleSet`: ns/op, B/op, allocs/op, MB/s. The orchestrator runs `go test -bench -benchmem` in each runner and produces one comparison table per ruleset with a Δ column (fork − upstream, negative = fork is faster/lighter).

### API contract parity (`results/contracts.md`)

Exported symbols in the analysis-related packages (`motor`, `rulesets`, `model`) are extracted from each runner via `golang.org/x/tools/go/packages`, then diffed. Output is a parity table plus a deterministic verdict:

- **drop-in** — symbol-for-symbol match; switching libraries needs no caller-code changes
- **minor-drift** — fork added symbols but existing callers unaffected
- **breaking-drift** — fork renamed/removed/re-signed symbols, OR the fork renamed its module path (both prevent drop-in swap)

## Layout

- `testdata/openapis/` — fetched OpenAPI specs (gitignored; run `make fetch`)
- `testdata/rulesets/` — five committed rulesets (recommended, all, owasp, minimal, big)
- `internal/scenarios/` — shared scenario matrix (no vacuum import)
- `runners/upstream/`, `runners/fork/` — per-version benchmark modules, each pinning its own vacuum
- `contracts/` — `golang.org/x/tools/go/packages`-backed API surface extractor + CLI
- `cmd/report/` — orchestrator + markdown renderer
- `scripts/fetch.sh` — SHA-pinned spec fetcher

## How module isolation works

The fork renamed its own module path — its `go.mod` declares `module github.com/buraksekili/vacuum`, not `github.com/daveshanley/vacuum`. That means the classic "swap via `replace`" pattern fails: Go refuses to let the same on-disk module serve under two identities (internal fork packages self-import at the new path). The fork runner therefore imports the fork directly at its actual path; bench and helper code differ from upstream's only in the two import lines.

The runners (`runners/upstream/`, `runners/fork/`) are **intentionally outside the Go workspace** (see `go.work` comment). If they were in the workspace, `go work sync` (or anything that implicitly triggers workspace-wide Minimum Version Selection) would rewrite the fork's pinned 2025-era pb33f deps to match upstream's 2026-era versions — invalidating the benchmark. All commands targeting runners use `GOWORK=off`. The Makefile handles this; you don't need to think about it.

**Never run `go work sync`** in this repository.

## Updating pinned versions

- **Upstream vacuum:** edit `runners/upstream/go.mod` `require` line, then `cd runners/upstream && GOWORK=off go mod tidy`.
- **Fork vacuum:** edit `runners/fork/go.mod` `require` line (the SHA), then `cd runners/fork && GOWORK=off go mod tidy`.
- **OpenAPI specs:** edit the three SHA constants at the top of `scripts/fetch.sh`, then re-run `make fetch`.

## Known limitations & caveats

- **Transitive-dep drift confounds performance comparisons.** The fork pins older pb33f/libopenapi (v0.23.0, July 2025), upstream has v0.36.1. A fork speedup on some scenarios may reflect faster old-deps as easily as genuine fork improvements; a fork slowdown may reflect older-deps regressions as easily as fork code drift. Treat per-scenario Δ figures as noisy signal, not clean measurement of fork-code-vs-upstream-code.
- **Fork fails on `$ref`-heavy specs.** The DigitalOcean API spec uses external `$ref`s. The fork's rolodex attempts to resolve them against a local path that doesn't exist in this project, erroring mid-bench. This surfaces as blank (`—`) cells in the `medium` row across rulesets. Upstream handles the same spec successfully. This is itself a drift finding, not a project bug.
- **No parse-vs-rules split.** Benchmarks measure end-to-end lint (parse + rule execution). If you want to attribute a delta to one phase, you'll need additional scenarios.
- **Only YAML specs are benchmarked.** JSON parsing cost is not measured.
- **Benchmark noise.** Laptop thermal state and background processes affect ns/op. Re-run with higher `COUNT` to reduce noise: `make bench COUNT=10` or higher.
- **Fork staleness (commits-behind-upstream) is out of scope.** This project measures runtime behaviour and API surface, not git history.

## Interpreting the contract report

Even when the verdict is `breaking-drift`, the parity table in `contracts.md` lists per-symbol status so you can see exactly what shifted. Typical shape for an older fork against a newer upstream:

- Many rows with status `match` — the common analysis API is stable.
- Rows with status `removed` — upstream has added symbols the fork doesn't have yet.
- Rows with status `added` — would indicate fork-only features; rare in this pairing.
- Rows with status `signature-changed` — would indicate a fork or upstream has rewritten a shared API; rare.

A `package`-kind row indicates the module path rename (fork is at `buraksekili/vacuum/motor`, upstream at `daveshanley/vacuum/motor`); the short-name keying in the diff still aligns them under `motor`.

## Reproducibility

- OpenAPI specs pinned by commit SHA in `scripts/fetch.sh`.
- vacuum versions pinned in each runner's `go.mod`.
- Go version declared in each module's `go.mod` (vacuum v0.26.1 requires Go 1.25).
- `make bench COUNT=N` records `count` in the report header.
- Report header records CPU model, Go version, and date.
