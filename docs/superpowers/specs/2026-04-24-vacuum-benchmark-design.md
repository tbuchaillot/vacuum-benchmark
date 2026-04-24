# Vacuum Fork Benchmark — Design

**Date:** 2026-04-24
**Author:** tbuchaillot
**Status:** Approved

## Goal

Build a reproducible comparison between two Go OpenAPI linter implementations:

- **Upstream:** `github.com/daveshanley/vacuum@v0.26.1`
- **Fork:** `github.com/buraksekili/vacuum` (HEAD)

The comparison answers two questions:

1. **Performance:** Across realistic spec sizes and rulesets, which linter is faster and lighter on allocations?
2. **Contract parity:** Do the two expose the same analysis-related public API? Can a consumer swap upstream for the fork via a `replace` directive, or are code changes required?

Fork staleness (git commits behind upstream) is **out of scope**.

## Non-Goals

- Correctness comparison of rule *results* (which violations each linter emits)
- Benchmarking CLI binaries — this measures Go library calls only
- Testing parse-only vs rules-only hot paths separately — every benchmark measures end-to-end lint
- Benchmarking JSON specs — YAML only, to avoid conflating format parsing with linting
- Measuring cold vs warm caches — Go's benchmark framework handles warm-up
- Comparing multiple fork revisions over time

## Approach Summary

Both libraries publish under the same Go import path (`github.com/daveshanley/vacuum`), so they cannot coexist in one binary. The solution is **two sub-modules inside a Go workspace**: one pins upstream via its `go.mod`, the other uses a `replace` directive to redirect the same import to the fork. A shared sub-module holds scenario metadata (no vacuum import, so it's safe for both). An orchestrator shells out to `go test -bench` in each runner, then generates a markdown report.

## Repository Layout

```
vacuum-benchmark/
├── go.work                              # stitches the sub-modules
├── Makefile                             # entry points: fetch, bench, contracts, all, clean
├── README.md
├── scripts/
│   └── fetch.sh                         # pins + downloads real-world OpenAPI specs
├── testdata/
│   ├── openapis/
│   │   ├── petstore.yaml                # small  (~180 lines)
│   │   ├── digitalocean.yaml            # medium (~3-5k lines)
│   │   └── github.yaml                  # large  (~40k lines)
│   └── rulesets/
│       ├── recommended.yaml             # extends vacuum built-in `recommended`
│       ├── all.yaml                     # extends vacuum built-in `all`
│       ├── owasp.yaml                   # extends vacuum built-in `owasp`
│       ├── minimal.yaml                 # 2-3 custom cheap rules
│       └── big.yaml                     # ~50-100 mix of built-in + custom rules
├── internal/
│   └── scenarios/                       # sub-module: shared scenario metadata
│       ├── go.mod                       # NO vacuum import
│       └── scenarios.go
├── runners/
│   ├── upstream/                        # sub-module: pins daveshanley/vacuum v0.26.1
│   │   ├── go.mod
│   │   └── bench_test.go
│   └── fork/                            # sub-module: replace → buraksekili/vacuum
│       ├── go.mod
│       └── bench_test.go
├── contracts/                           # sub-module: API surface extraction
│   ├── go.mod
│   └── cmd/contracts/main.go            # dumps exported analysis symbols → JSON
├── cmd/
│   └── report/                          # sub-module: orchestrator
│       ├── go.mod
│       └── main.go                      # runs benches+contracts, writes markdown
└── results/                             # generated outputs (gitignored)
    ├── results.md
    ├── contracts.md
    └── raw/                             # per-runner bench stdout + upstream.json + fork.json
```

## Scenario Model

The shared module `internal/scenarios` defines scenarios as pure data. It **must not import vacuum** — that's what lets both runners depend on it without module conflict.

```go
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
    SpecPath    string // e.g. "testdata/openapis/petstore.yaml", relative to repo root
    RulesetPath string // e.g. "testdata/rulesets/recommended.yaml", relative to repo root
}

// All returns the cartesian product: 3 sizes × 5 rulesets = 15 scenarios.
func All() []Scenario
```

**Repo-root resolution:** scenario paths are relative to the repo root. Each runner's `bench_test.go` calls a helper that walks up from the test binary's working directory until it finds `go.work`, then joins the scenario path.

**Benchmark naming:** each scenario becomes a sub-benchmark named `BenchmarkLint/<size>/<ruleset>`, producing 15 measurements per runner (30 total).

## Benchmark Runners

Both runners are structurally identical. Only the `go.mod` differs.

**`runners/upstream/go.mod`:**
```
module vacuum-benchmark/runners/upstream

go 1.22

require github.com/daveshanley/vacuum v0.26.1
require vacuum-benchmark/internal/scenarios v0.0.0
```

**`runners/fork/go.mod`:**
```
module vacuum-benchmark/runners/fork

go 1.22

require github.com/daveshanley/vacuum v0.0.0
require vacuum-benchmark/internal/scenarios v0.0.0

replace github.com/daveshanley/vacuum => github.com/buraksekili/vacuum <pinned-rev>
```

Both runners share this benchmark body:

```go
package bench_test

import (
    "os"
    "testing"

    "github.com/daveshanley/vacuum/motor"
    "github.com/daveshanley/vacuum/rulesets"

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
                    b.Fatal("nil result")
                }
            }
        })
    }
}
```

**What's timed:** full `ApplyRulesToRuleSet` — parse plus rule execution. This is what consumers actually call.

**What's outside the timed loop:** reading the spec file and constructing the ruleset object. These are one-time costs, not lint hot path.

**Metrics captured:** ns/op, B/op, allocs/op (via `ReportAllocs`) and MB/s (via `SetBytes`).

**Runner build failure:** if the fork has diverged on `motor.ApplyRulesToRuleSet` or `RuleSetExecution`, the fork runner won't compile. That is itself a finding — the Makefile records the build status per runner and continues. The contract report surfaces the exact symbol mismatch.

## Contract Extraction & Diff

**Scope:** analysis-related symbols only — the packages a consumer touches when linting.

- `github.com/daveshanley/vacuum/motor`
- `github.com/daveshanley/vacuum/rulesets`
- `github.com/daveshanley/vacuum/model`

**Extraction (`contracts/cmd/contracts`):**

Uses `golang.org/x/tools/go/packages` to load the target module. The orchestrator invokes it twice — once with cwd set to `runners/upstream`, once with cwd set to `runners/fork` — so each invocation sees that runner's resolved vacuum version.

Output: structured JSON per run.

```json
{
  "version": "upstream-v0.26.1",
  "packages": {
    "github.com/daveshanley/vacuum/motor": {
      "functions": [
        {"name": "ApplyRulesToRuleSet", "signature": "func(execution *RuleSetExecution) *RuleSetExecutionResult"}
      ],
      "types": [
        {"name": "RuleSetExecution", "kind": "struct", "fields": [
          {"name": "RuleSet", "type": "*rulesets.RuleSet"},
          {"name": "Spec", "type": "[]byte"}
        ]}
      ],
      "methods": [...]
    }
  }
}
```

**Diff:**

The report generator reads `results/raw/upstream.json` and `results/raw/fork.json` and produces `results/contracts.md`:

1. **Parity table** — one row per symbol, columns `upstream` / `fork` / `status` where status is one of:
   - `match` — signature identical
   - `signature-changed` — present in both, signatures differ
   - `removed` — in upstream, absent from fork
   - `added` — in fork, absent from upstream
2. **Drift-risk verdict** — a short deterministic paragraph classifying the fork:
   - **Drop-in** — all symbols match; swap via `replace`, no consumer changes
   - **Minor drift** — only `added` entries; still drop-in for existing callers
   - **Breaking drift** — any `signature-changed` or `removed` entries; consumers must update

The verdict is computed from the diff, not hand-written.

**Why JSON-in-the-middle:** separating extraction from diffing means each runner's module stays self-contained (extraction only needs its runner's imports) and the diff is re-runnable without reloading modules.

## Report Generation & Makefile

**`cmd/report`** orchestrates. It does not import vacuum.

**Flow:**

1. For each runner (`upstream`, `fork`):
   - Run `go test -bench=BenchmarkLint -benchmem -run=^$ -count=$(COUNT) ./...` with cwd set to that runner's directory.
   - Parse stdout: each `BenchmarkLint/<size>/<ruleset>-<GOMAXPROCS>` line yields (iterations, ns/op, B/op, allocs/op, MB/s).
   - If the runner fails to build (compiler error), record `BUILD_FAILED` with the error text and continue.
2. Invoke the contracts binary once per runner → `results/raw/upstream.json`, `results/raw/fork.json`.
3. Diff the two JSONs.
4. Write:
   - `results/results.md` — per-ruleset tables; rows = sizes; columns = upstream/fork for ns/op, B/op, allocs/op; plus a `Δ%` column. Top of file: build status per runner, Go version, CPU model, date, COUNT used.
   - `results/contracts.md` — parity table + drift-risk verdict.

**Sample `results.md` fragment:**

```
## Ruleset: recommended

| Size   | upstream ns/op | fork ns/op | Δ       | upstream B/op | fork B/op | Δ       | upstream allocs | fork allocs | Δ       |
|--------|----------------|------------|---------|---------------|-----------|---------|-----------------|-------------|---------|
| small  | 1,234,567      | 1,100,000  | -10.9%  | 350 KB        | 320 KB    | -8.6%   | 4,821           | 4,500       | -6.7%   |
| medium | ...            | ...        | ...     | ...           | ...       | ...     | ...             | ...         | ...     |
```

Δ is `(fork - upstream) / upstream * 100` — negative means fork is faster/lighter.

**Makefile targets:**

```
make fetch       # download+pin real-world specs into testdata/openapis/
make bench       # run benchmarks in both runners, regenerate results.md
make contracts   # extract+diff API surface, regenerate contracts.md
make all         # fetch + bench + contracts
make clean       # wipe results/
```

`make bench` honours `COUNT=<n>` (default 5) to tune statistical noise.

**Why shell out to `go test` instead of using `testing.Benchmark` programmatically:** running each runner in its own `go test` process guarantees clean module resolution and prevents any chance of one runner's transitive deps leaking into the other.

## Test Data Strategy

### OpenAPI specs (`testdata/openapis/`)

Fetched by `make fetch` (runs `scripts/fetch.sh`) — each URL pinned to a specific commit for reproducibility.

| Size   | Source                                                                              | Approx lines | License    |
|--------|-------------------------------------------------------------------------------------|--------------|------------|
| small  | Swagger Petstore v3 from `OAI/OpenAPI-Specification` (pinned commit)                | ~180         | Apache-2.0 |
| medium | DigitalOcean API from APIs.guru registry (pinned revision)                          | ~3-5k        | Apache-2.0 |
| large  | GitHub REST API (bundled spec) from `github/rest-api-description` (pinned commit)   | ~40k         | MIT        |

**Why not Stripe for "large":** ~400k lines; a single bench run would take 10+ minutes per size×ruleset, drowning signal in wall-clock cost. GitHub at ~40k is large enough to differ meaningfully from medium while finishing in reasonable time.

**Fetch over commit:** specs are gitignored and downloaded on demand. Keeps the repo small and makes provenance obvious. `scripts/fetch.sh` is idempotent and pins by SHA.

### Rulesets (`testdata/rulesets/`)

All five are committed to the repo (they are small).

| File             | Contents                                                                              |
|------------------|---------------------------------------------------------------------------------------|
| recommended.yaml | Extends vacuum's built-in `recommended` ruleset                                       |
| all.yaml         | Extends vacuum's built-in `all` ruleset                                               |
| owasp.yaml       | Extends vacuum's built-in `owasp` ruleset                                             |
| minimal.yaml     | 2-3 cheap rules (e.g., `info-contact`, `info-license`, `openapi-tags`)                |
| big.yaml         | ~50-100 rules: mix of built-in + custom. Isolates rule-count overhead from engine overhead |

Uses vacuum's native ruleset YAML schema. The fork preserves this schema; if it has diverged, the failure shows up as a build/load error in one runner and is reported by the orchestrator.

## Reproducibility

- Every external asset is pinned by commit SHA.
- Go versions are pinned in each `go.mod` via `go 1.22`.
- `go.work` locks the sub-module set.
- `make bench COUNT=N` documents the statistical sample size in the report header.
- Report header records CPU model and Go version so cross-machine comparisons stay honest.

## Risks & Open Questions

- **Fork API drift** — if the fork has renamed `motor.ApplyRulesToRuleSet` or changed `RuleSetExecution`, the fork runner won't compile. Handled: Makefile reports per-runner build status; contracts report pinpoints the symbol that moved.
- **Ruleset YAML drift** — if the fork changed the ruleset schema, `buildRuleset` will fail at runtime. Handled: surfaced as a bench failure in the report.
- **Benchmark noise** — laptop thermal state, background processes. Partially mitigated by `count=5` default and reporting the CPU model. Not fully solved; not in scope to solve.
- **Fetch-at-runtime** — if an upstream spec moves or gets deleted, `make fetch` breaks. Accepted trade-off vs. committing large specs.

## Deliverables

1. A runnable `make all` that produces `results/results.md` and `results/contracts.md` on a clean checkout.
2. All five rulesets committed.
3. `scripts/fetch.sh` pinning the three OpenAPI specs by commit SHA.
4. README documenting: how to run, how to interpret results, known limitations.
