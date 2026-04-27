# vacuum-benchmark

A reproducible Go benchmark comparing two OpenAPI linter implementations:

- **Upstream:** [`github.com/daveshanley/vacuum`](https://github.com/daveshanley/vacuum) pinned to `v0.26.1`
- **Fork:** [`github.com/buraksekili/vacuum`](https://github.com/buraksekili/vacuum) pinned to commit `b857c8f` (2025-07)

It measures end-to-end lint performance across spec size × ruleset combinations and diffs the analysis-related public API surface.

Design and plan documents live under `docs/superpowers/`.

## TL;DR — findings on Apple M5 Pro

*From a `make bench COUNT=2` run; signs are noisy at low counts but the pattern is consistent. Full reports: [`results/results.md`](results/results.md) (per-ruleset perf tables) and [`results/contracts.md`](results/contracts.md) (API parity + verdict).*

- **Verdict: `breaking-drift`.** The fork can't be swapped in via Go's `replace` directive; it renamed its own module path. Several `rulesets` and `model` symbols upstream callers may use are absent from the fork.
- **Speed:** upstream wins on medium and large specs. Fork is +50–140% slower on large; on small with light rulesets fork is roughly +90% slower.
- **Memory:** fork uses ~30% less memory and ~14% fewer allocs for *small + heavy ruleset*. Everywhere else it allocates more — ~2× memory and ~2.6× allocs on large.
- **Fork crashes on the medium spec** (DigitalOcean — uses external `$ref`s the fork can't resolve in this layout). Medium rows render as `—`.

Caveat: the fork pins 2025-era pb33f deps (`libopenapi v0.23` vs upstream's `v0.36`), so an unknown share of the perf delta is dep-version drift rather than vacuum-code drift. See [Known limitations](#known-limitations--caveats).

## Quick Start

```bash
make fetch      # downloads pinned OpenAPI specs into testdata/openapis/
make bench      # runs benchmarks + contract diff; writes results/results.md + results/contracts.md
```

For more statistical stability, raise the iteration count:

```bash
make bench COUNT=10
```

`results/results.md` is the perf table; `results/contracts.md` is the API parity table + verdict. Both are gitignored — regenerate any time.

## Runtime expectations

`make bench` is dominated by linting the large GitHub spec (~9 MB, multiple seconds per iteration). On Apple M5 Pro:

| `COUNT` | Wall-clock |
|---|---|
| 1 | ~1m 40s |
| 2 | ~2m |
| 5 (default) | ~4–5m |
| 10 | ~8–10m |

Per-iteration cost of a single `motor.ApplyRulesToRuleSet`:

- **small** (Petstore, ~3 KB): 1–4 ms upstream, 2–5 ms fork
- **medium** (DigitalOcean, ~100 KB): 15–525 ms upstream; fork errors out
- **large** (GitHub, ~9 MB): 2–6 s upstream, 4.5–9 s fork

For fast iteration during dev, run a subset directly:

```bash
cd runners/upstream && GOWORK=off go test -bench=BenchmarkLint/small -benchtime=1x ./...
```

That executes all 5 small-size scenarios in ~10 seconds.

## What gets measured

### Performance (`results/results.md`)

For each of 3 spec sizes × 5 rulesets = 15 scenarios, `BenchmarkLint` measures end-to-end `motor.ApplyRulesToRuleSet`: ns/op, B/op, allocs/op, MB/s. The orchestrator runs `go test -bench -benchmem` in each runner and produces one comparison table per ruleset with a Δ column.

Δ is computed as `(fork − upstream) / upstream × 100`. **Negative Δ means the fork is faster/lighter; positive means slower/heavier.**

### API contract parity (`results/contracts.md`)

Exported symbols in the analysis-related packages (`motor`, `rulesets`, `model`) are extracted from each runner via `golang.org/x/tools/go/packages`, then diffed. The parity table lists every symbol with status `match` / `added` / `removed` / `signature-changed`. The header carries one of three deterministic verdicts:

- **drop-in** — symbol-for-symbol match; switching libraries needs no caller-code changes.
- **minor-drift** — fork added symbols but existing callers remain source-compatible.
- **breaking-drift** — fork renamed/removed/re-signed symbols, OR the fork renamed its module path (both prevent drop-in swap).

## Layout

- `testdata/openapis/` — fetched OpenAPI specs (gitignored; run `make fetch`)
- `testdata/rulesets/` — five committed rulesets (recommended, all, owasp, minimal, big)
- `internal/scenarios/` — shared scenario matrix; no vacuum import
- `runners/upstream/`, `runners/fork/` — per-version benchmark modules, each pinning its own vacuum
- `contracts/` — `go/packages`-backed API surface extractor + CLI
- `cmd/report/` — orchestrator + benchparse + contractdiff + markdown renderer
- `scripts/fetch.sh` — SHA-pinned spec fetcher
- `Makefile` — top-level entry points (`fetch`, `bench`, `test`, `tidy`, `clean`)

## Targets

| Target | What it does |
|---|---|
| `make fetch` | Downloads SHA-pinned OpenAPI specs into `testdata/openapis/` (idempotent). |
| `make bench [COUNT=N]` | Runs benchmarks in both runners + contract diff; writes `results/results.md` and `results/contracts.md`. Default `COUNT=5`. |
| `make all` | Equivalent to `make fetch bench`. |
| `make test` | Runs unit tests across all sub-modules. |
| `make tidy` | `go mod tidy` per sub-module (with `GOWORK=off` for runners). |
| `make clean` | `rm -rf results/`. |

## How module isolation works

The fork renamed its own module path — its `go.mod` declares `module github.com/buraksekili/vacuum`, not `github.com/daveshanley/vacuum`. That breaks the standard "swap via `replace`" pattern: Go refuses to let the same on-disk module serve under two identities (internal fork packages self-import at the new path). The fork runner therefore imports the fork directly at its actual path; bench and helper code differ from upstream's only in two import lines.

The runners (`runners/upstream/`, `runners/fork/`) are **intentionally outside the Go workspace** (see the comment in `go.work`). If they were in the workspace, `go work sync` (or anything that implicitly triggers workspace-wide Minimum Version Selection) would rewrite the fork's pinned 2025-era pb33f deps to match upstream's 2026-era versions — invalidating the benchmark. All commands targeting runners use `GOWORK=off`; the Makefile handles this for you.

> **Never run `go work sync` in this repository.** It will silently destroy the fork's pinned dep graph.

## Updating pinned versions

- **Upstream vacuum:** edit `runners/upstream/go.mod` `require` line, then `cd runners/upstream && GOWORK=off go mod tidy`.
- **Fork vacuum:** edit `runners/fork/go.mod` `require` line (the SHA), then `cd runners/fork && GOWORK=off go mod tidy`.
- **OpenAPI specs:** edit the three SHA constants at the top of `scripts/fetch.sh`, then re-run `make fetch`.

After bumping any pin, regenerate the reports with `make bench`.

## Interpreting the contract report

Even when the verdict is `breaking-drift`, the parity table in `contracts.md` lists per-symbol status so you can see exactly what shifted. Typical shape for an older fork against a newer upstream:

- **`match`** — the common analysis API is stable. Most rows.
- **`removed`** — upstream has added symbols the fork doesn't have yet.
- **`added`** — fork-only features; rare in this pairing.
- **`signature-changed`** — fork or upstream rewrote a shared API. Rare; worth investigating closely when present.

A `package`-kind row indicates the module path rename. The diff keys packages by short name (`motor`, `rulesets`, `model`), so upstream's `github.com/daveshanley/vacuum/motor` and fork's `github.com/buraksekili/vacuum/motor` line up under one row in the parity table.

## Known limitations & caveats

- **Transitive-dep drift confounds performance comparisons.** The fork pins older pb33f/libopenapi (v0.23.0, July 2025); upstream is at v0.36.1. A fork speedup may reflect faster old-deps as easily as fork code improvements; a fork slowdown may reflect older-dep regressions. Treat per-scenario Δ figures as directional signal, not clean measurement of fork-code-vs-upstream-code.
- **Fork crashes on `$ref`-heavy specs.** The DigitalOcean API spec uses external `$ref`s. The fork's rolodex tries to resolve them against a local path that doesn't exist in this project, erroring mid-bench. The medium row therefore renders as `—` for the fork. Upstream handles the same spec successfully — this is a real drift finding, not a project bug.
- **No parse-vs-rules split.** Benchmarks measure end-to-end lint (parse + rule execution). Attributing a delta to one phase requires additional scenarios.
- **Only YAML specs are benchmarked.** JSON parsing cost is not measured.
- **Benchmark noise.** Laptop thermal state and background processes affect ns/op. Re-run with higher `COUNT` to reduce noise.
- **Fork staleness (commits-behind-upstream) is out of scope.** This project measures runtime behaviour and API surface, not git history.

## Reproducibility

- OpenAPI specs pinned by commit SHA in `scripts/fetch.sh`.
- vacuum versions pinned in each runner's `go.mod` (upstream tag, fork SHA).
- Go directive declared per module (vacuum v0.26.1 requires Go 1.25).
- `make bench COUNT=N` records `count` in the report header.
- Report header records CPU model, Go version, and date.
- Raw bench stdout/stderr are persisted in `results/raw/` for post-mortem inspection.
