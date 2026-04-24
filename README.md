# vacuum-benchmark

A reproducible benchmark comparing two Go OpenAPI linter implementations:

- **Upstream:** [`github.com/daveshanley/vacuum`](https://github.com/daveshanley/vacuum) pinned to `v0.26.1`
- **Fork:** [`github.com/buraksekili/vacuum`](https://github.com/buraksekili/vacuum) at HEAD

Measures lint performance (ns/op, B/op, allocs/op) across spec sizes × rulesets, and diffs the analysis-related public API surface to classify drift.

See `docs/superpowers/specs/2026-04-24-vacuum-benchmark-design.md` for the design.

## Quick Start

```bash
make fetch      # download real-world OpenAPI specs into testdata/openapis/
make all        # run benchmarks + contract diff; writes results/results.md + results/contracts.md
```

## Layout

- `testdata/openapis/` — fetched real-world OpenAPI specs (small/medium/large)
- `testdata/rulesets/` — committed rulesets (recommended, all, owasp, minimal, big)
- `runners/upstream/`, `runners/fork/` — per-version benchmark modules
- `internal/scenarios/` — shared scenario metadata (no vacuum import)
- `contracts/` — API surface extractor (uses `golang.org/x/tools/go/packages`)
- `cmd/report/` — orchestrator; shells out to `go test -bench` and generates markdown reports
- `scripts/fetch.sh` — SHA-pinned fetch of upstream OpenAPI specs
