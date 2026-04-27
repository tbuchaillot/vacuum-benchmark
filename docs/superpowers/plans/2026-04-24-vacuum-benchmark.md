# Vacuum Fork Benchmark Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a reproducible Go benchmark comparing `github.com/buraksekili/vacuum` against `github.com/daveshanley/vacuum@v0.26.1` on lint performance across spec sizes and rulesets, plus an analysis-API contract parity report.

**Architecture:** Go workspace with four sub-modules — `runners/upstream` (pins v0.26.1), `runners/fork` (uses `replace` to redirect to fork), `internal/scenarios` (pure-data scenario list, no vacuum import), `contracts` (go/packages symbol extractor), `cmd/report` (orchestrator that shells out to `go test -bench`, runs the contracts extractor per runner, and emits `results.md` + `contracts.md`).

**Tech Stack:** Go 1.22, Go workspaces (`go.work`), `testing.B` sub-benchmarks, `golang.org/x/tools/go/packages` for API extraction, Make, bash for fetch script, vacuum's native ruleset YAML.

**Spec:** `docs/superpowers/specs/2026-04-24-vacuum-benchmark-design.md`

---

## Conventions

- All paths are **relative to the repo root** `/Users/tombuchaillot/conductor/workspaces/vacuum-benchmark/milan`.
- All Go files use `go 1.22`.
- All runners declare `module vacuum-benchmark/<path>` — the workspace resolves these locally.
- Commit messages use Conventional Commits (`feat:`, `chore:`, `test:`, `docs:`).
- After each task's final step, run `git status` to confirm a clean working tree before moving on.

---

## Task 1: Repo Skeleton and Workspace File

**Files:**
- Create: `go.work`
- Create: `.gitignore`
- Modify: `README.md` (replace 1-line stub with project overview)

- [ ] **Step 1: Create `.gitignore`**

Write `.gitignore`:

```gitignore
# Generated benchmark outputs
/results/

# Fetched OpenAPI specs (downloaded by `make fetch`)
/testdata/openapis/*.yaml
/testdata/openapis/*.json

# Go build artifacts
*.test
*.out

# Editor
.DS_Store
.idea/
.vscode/
```

- [ ] **Step 2: Create `go.work` skeleton**

Write `go.work`:

```
go 1.22

use (
    ./internal/scenarios
    ./runners/upstream
    ./runners/fork
    ./contracts
    ./cmd/report
)
```

This file will error until the sub-modules exist. That's fine — we'll create them in subsequent tasks. We commit the workspace file now so later `go.mod` init commands see the expected structure.

- [ ] **Step 3: Update README with project overview**

Overwrite `README.md`:

```markdown
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
```

- [ ] **Step 4: Commit**

```bash
git add .gitignore go.work README.md
git commit -m "chore: add workspace skeleton and .gitignore"
```

Expected: one commit created, working tree clean.

---

## Task 2: Committed Rulesets

**Files:**
- Create: `testdata/rulesets/recommended.yaml`
- Create: `testdata/rulesets/all.yaml`
- Create: `testdata/rulesets/owasp.yaml`
- Create: `testdata/rulesets/minimal.yaml`
- Create: `testdata/rulesets/big.yaml`

These are data files; no test step. They will be exercised by the benchmarks in later tasks.

- [ ] **Step 1: Create `testdata/rulesets/recommended.yaml`**

```yaml
extends: [[spectral:oas, recommended]]
```

- [ ] **Step 2: Create `testdata/rulesets/all.yaml`**

```yaml
extends: [[spectral:oas, all]]
```

- [ ] **Step 3: Create `testdata/rulesets/owasp.yaml`**

```yaml
extends: [[spectral:oas, off]]
rules:
  owasp-no-numeric-ids: true
  owasp-no-http-basic: true
  owasp-no-api-keys-in-url: true
  owasp-no-credentials-in-url: true
  owasp-auth-insecure-schemes: true
  owasp-jwt-best-practices: true
  owasp-protection-global-unsafe: true
  owasp-protection-global-unsafe-strict: true
  owasp-protection-global-safe: true
  owasp-define-error-validation: true
  owasp-define-error-responses-401: true
  owasp-define-error-responses-500: true
  owasp-rate-limit: true
  owasp-rate-limit-retry-after: true
  owasp-rate-limit-responses-429: true
  owasp-define-error-responses-429: true
  owasp-array-limit: true
  owasp-string-limit: true
  owasp-string-restricted: true
  owasp-integer-limit: true
  owasp-integer-format: true
  owasp-no-additionalProperties: true
  owasp-constrained-additionalProperties: true
  owasp-security-hosts-https-oas3: true
```

- [ ] **Step 4: Create `testdata/rulesets/minimal.yaml`**

```yaml
extends: [[spectral:oas, off]]
rules:
  info-contact: true
  info-license: true
  openapi-tags: true
```

- [ ] **Step 5: Create `testdata/rulesets/big.yaml`**

```yaml
extends: [[spectral:oas, all]]
rules:
  owasp-no-numeric-ids: true
  owasp-no-http-basic: true
  owasp-no-api-keys-in-url: true
  owasp-no-credentials-in-url: true
  owasp-auth-insecure-schemes: true
  owasp-jwt-best-practices: true
  owasp-protection-global-unsafe: true
  owasp-protection-global-safe: true
  owasp-define-error-validation: true
  owasp-define-error-responses-401: true
  owasp-define-error-responses-500: true
  owasp-rate-limit: true
  owasp-rate-limit-responses-429: true
  owasp-array-limit: true
  owasp-string-limit: true
  owasp-integer-limit: true
  owasp-no-additionalProperties: true
```

- [ ] **Step 6: Commit**

```bash
git add testdata/rulesets/
git commit -m "feat: add five test rulesets (recommended, all, owasp, minimal, big)"
```

Expected: working tree clean.

---

## Task 3: Fetch Script for OpenAPI Specs

**Files:**
- Create: `scripts/fetch.sh`

`fetch.sh` downloads three specs, each pinned by commit SHA. The plan resolves SHAs at implementation time via `gh api`; the resolved script is what we commit.

- [ ] **Step 1: Resolve pinned SHAs**

Run these three commands and capture the `.sha` field of each response. Record the three SHAs for use in the script below.

```bash
gh api repos/OAI/OpenAPI-Specification/commits/main --jq '.sha'
gh api repos/digitalocean/openapi/commits/main --jq '.sha'
gh api repos/github/rest-api-description/commits/main --jq '.sha'
```

Expected: three 40-character hex strings (e.g., `3b9f3f1...`). If `gh` is not authenticated, run `gh auth login` first or fall back to `curl -s https://api.github.com/repos/<owner>/<repo>/commits/main | jq -r .sha`.

Verify each spec URL resolves by `curl -sI`:

```bash
curl -sI "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/<PETSTORE_SHA>/examples/v3.0/petstore.yaml" | head -1
curl -sI "https://raw.githubusercontent.com/digitalocean/openapi/<DO_SHA>/specification/DigitalOcean-public.v2.yaml" | head -1
curl -sI "https://raw.githubusercontent.com/github/rest-api-description/<GITHUB_SHA>/descriptions/api.github.com/api.github.com.yaml" | head -1
```

Expected: each returns `HTTP/2 200`. If any returns 404, the path has moved — locate the correct path in that repo (via `gh api repos/<owner>/<repo>/contents/<dir>`) and substitute below.

- [ ] **Step 2: Write `scripts/fetch.sh` with resolved SHAs**

Substitute `<PETSTORE_SHA>`, `<DO_SHA>`, `<GITHUB_SHA>` with the 40-character values from Step 1. Replace the two `<...>` path placeholders with the confirmed paths from Step 1.

```bash
#!/usr/bin/env bash
set -euo pipefail

# SHA-pinned OpenAPI spec fetch. Idempotent.
# Edit the SHAs below to update; re-run `make fetch`.

PETSTORE_SHA="<PETSTORE_SHA>"
DO_SHA="<DO_SHA>"
GITHUB_SHA="<GITHUB_SHA>"

PETSTORE_URL="https://raw.githubusercontent.com/OAI/OpenAPI-Specification/${PETSTORE_SHA}/examples/v3.0/petstore.yaml"
DO_URL="https://raw.githubusercontent.com/digitalocean/openapi/${DO_SHA}/specification/DigitalOcean-public.v2.yaml"
GITHUB_URL="https://raw.githubusercontent.com/github/rest-api-description/${GITHUB_SHA}/descriptions/api.github.com/api.github.com.yaml"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${REPO_ROOT}/testdata/openapis"
mkdir -p "${OUT}"

fetch() {
    local url="$1"
    local dest="$2"
    echo "fetching ${url} -> ${dest}"
    curl --fail --silent --show-error --location -o "${dest}" "${url}"
    echo "  size: $(wc -c < "${dest}") bytes"
}

fetch "${PETSTORE_URL}" "${OUT}/petstore.yaml"
fetch "${DO_URL}"       "${OUT}/digitalocean.yaml"
fetch "${GITHUB_URL}"   "${OUT}/github.yaml"

echo "done. fetched specs:"
ls -la "${OUT}"
```

- [ ] **Step 3: Make executable and smoke test**

```bash
chmod +x scripts/fetch.sh
./scripts/fetch.sh
```

Expected: three files in `testdata/openapis/` — `petstore.yaml` (~3-10 KB), `digitalocean.yaml` (~several hundred KB), `github.yaml` (~several MB). If the medium file is over 2 MB, that's fine; if it's over 10 MB the medium choice will drown out the large signal — flag and adjust the URL.

- [ ] **Step 4: Commit**

```bash
git add scripts/fetch.sh
git commit -m "feat: add SHA-pinned fetch script for openapi specs"
```

The downloaded `.yaml` files stay uncommitted (gitignored by Task 1).

---

## Task 4: `internal/scenarios` Module

**Files:**
- Create: `internal/scenarios/go.mod`
- Create: `internal/scenarios/scenarios.go`
- Create: `internal/scenarios/scenarios_test.go`

This module is the single source of truth for what benchmarks run. It must not import vacuum.

- [ ] **Step 1: Initialise the module**

```bash
cd internal/scenarios
go mod init vacuum-benchmark/internal/scenarios
cd ../..
```

Edit `internal/scenarios/go.mod` to confirm the `go 1.22` directive:

```
module vacuum-benchmark/internal/scenarios

go 1.22
```

- [ ] **Step 2: Write the failing test**

Create `internal/scenarios/scenarios_test.go`:

```go
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
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd internal/scenarios && go test ./... ; cd ../..
```

Expected: compile error — `undefined: All`, `undefined: Size`, etc.

- [ ] **Step 4: Implement `scenarios.go`**

Create `internal/scenarios/scenarios.go`:

```go
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
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd internal/scenarios && go test ./... ; cd ../..
```

Expected: `ok vacuum-benchmark/internal/scenarios`.

- [ ] **Step 6: Commit**

```bash
git add internal/scenarios/
git commit -m "feat: add internal/scenarios module defining bench matrix"
```

---

## Task 5: `runners/upstream` — Helpers

**Files:**
- Create: `runners/upstream/go.mod`
- Create: `runners/upstream/helpers_test.go` (test helpers used by bench, kept in `_test.go` so they don't ship as library code)

- [ ] **Step 1: Initialise module**

```bash
mkdir -p runners/upstream
cd runners/upstream
go mod init vacuum-benchmark/runners/upstream
go mod edit -require=github.com/daveshanley/vacuum@v0.26.1
go mod edit -require=vacuum-benchmark/internal/scenarios@v0.0.0
cd ../..
```

Confirm `runners/upstream/go.mod` looks like:

```
module vacuum-benchmark/runners/upstream

go 1.22

require (
	github.com/daveshanley/vacuum v0.26.1
	vacuum-benchmark/internal/scenarios v0.0.0
)
```

- [ ] **Step 2: Run `go mod tidy` against the workspace**

```bash
cd runners/upstream && go mod tidy ; cd ../..
```

Expected: the transitive deps of `daveshanley/vacuum@v0.26.1` are added to `go.sum`. Go resolves `vacuum-benchmark/internal/scenarios` from the workspace (no download). If `go mod tidy` complains that it cannot find the scenarios module, re-verify `go.work` is committed from Task 1 and that the cwd is inside the module's directory.

- [ ] **Step 3: Write the helpers**

Create `runners/upstream/helpers_test.go`:

```go
package bench_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daveshanley/vacuum/rulesets"
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
```

- [ ] **Step 4: Write a unit test for the helpers**

Append to `runners/upstream/helpers_test.go`:

```go
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
```

- [ ] **Step 5: Run tests**

```bash
cd runners/upstream && go test ./... ; cd ../..
```

Expected: `TestRepoRootFindsGoWork`, `TestBuildRulesetParsesRecommended` PASS. `TestMustReadFileReadsPetstore` may SKIP if specs aren't fetched. If `TestBuildRulesetParsesRecommended` fails with an API error on `rulesets.BuildDefaultRuleSets` or `GenerateRuleSetFromSuppliedRuleSet`, the upstream API in v0.26.1 differs from what we assumed — open `~/go/pkg/mod/github.com/daveshanley/vacuum@v0.26.1/rulesets/` and adjust helper calls to match. Common alternative names: `BuildDefaultRuleSets()`, `NewRuleSets()`, `CreateRuleSetFromData()`.

- [ ] **Step 6: Commit**

```bash
git add runners/upstream/go.mod runners/upstream/go.sum runners/upstream/helpers_test.go
git commit -m "feat(upstream): add module skeleton and bench helpers"
```

---

## Task 6: `runners/upstream` — Benchmark

**Files:**
- Create: `runners/upstream/bench_test.go`

- [ ] **Step 1: Write the benchmark**

Create `runners/upstream/bench_test.go`:

```go
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
```

- [ ] **Step 2: Compile-check by running with a short timeout**

```bash
cd runners/upstream && go test -run=^$ -bench=BenchmarkLint/small/minimal -benchtime=1x ./... ; cd ../..
```

Expected: one line like `BenchmarkLint/small/minimal-<N>   1   <time> ns/op   <bytes> B/op   <allocs> allocs/op`.

If `motor.ApplyRulesToRuleSet` or `motor.RuleSetExecution` don't exist in v0.26.1, open the package source and adapt. This is also a signal worth noting — the upstream runner should always compile cleanly against its pinned version.

If testdata isn't fetched, this will fail with `read testdata/openapis/petstore.yaml`. Run `make fetch` first (Task 14 wires this up, but at this point fetch.sh from Task 3 works standalone: `./scripts/fetch.sh`).

- [ ] **Step 3: Commit**

```bash
git add runners/upstream/bench_test.go
git commit -m "feat(upstream): add BenchmarkLint over scenario matrix"
```

---

## Task 7: `runners/fork` — Mirror with Replace Directive

**Files:**
- Create: `runners/fork/go.mod`
- Create: `runners/fork/helpers_test.go` (byte-for-byte copy of upstream's)
- Create: `runners/fork/bench_test.go` (byte-for-byte copy of upstream's)

- [ ] **Step 1: Resolve the fork's pinned revision**

```bash
gh api repos/buraksekili/vacuum/commits/main --jq '.sha'
```

Expected: a 40-character SHA. Record as `<FORK_SHA>` for substitution below.

- [ ] **Step 2: Initialise fork module**

```bash
mkdir -p runners/fork
cd runners/fork
go mod init vacuum-benchmark/runners/fork
cd ../..
```

Write `runners/fork/go.mod` (substitute `<FORK_SHA>`):

```
module vacuum-benchmark/runners/fork

go 1.22

require (
	github.com/daveshanley/vacuum v0.0.0
	vacuum-benchmark/internal/scenarios v0.0.0
)

replace github.com/daveshanley/vacuum => github.com/buraksekili/vacuum <FORK_SHA>
```

Note: for a Git commit reference (not a tagged version) the `replace` right-hand side uses just the SHA (Go modules resolves it to a pseudo-version).

- [ ] **Step 3: Resolve the replacement**

```bash
cd runners/fork && go mod tidy ; cd ../..
```

Expected: `go.sum` populated. If `go mod tidy` fails because the fork's module path differs (e.g., the fork declares a different `module` line in its own `go.mod`), the replace target needs to match. Inspect the fork's go.mod via `gh api repos/buraksekili/vacuum/contents/go.mod --jq '.content' | base64 -d | head -5`. If its module path is e.g. `github.com/buraksekili/vacuum` — this is the expected case. If it still declares `github.com/daveshanley/vacuum` (typical for straight forks), `go mod tidy` will succeed.

If the fork declares a different module path, adjust the replace's **left side** to match consumers' imports: keep `github.com/daveshanley/vacuum` on the left (since that's what the bench code imports) and the fork's module path on the right. That's already the case in Step 2.

- [ ] **Step 4: Copy helpers and bench from upstream**

```bash
cp runners/upstream/helpers_test.go runners/fork/helpers_test.go
cp runners/upstream/bench_test.go   runners/fork/bench_test.go
```

No content changes — the fork preserves the same import path, so the same source compiles against both runners.

- [ ] **Step 5: Compile-check**

```bash
cd runners/fork && go test -run=^$ -bench=BenchmarkLint/small/minimal -benchtime=1x ./... ; cd ../..
```

Expected: a benchmark line. If compilation fails because the fork has renamed `motor.ApplyRulesToRuleSet`, `motor.RuleSetExecution`, or any rulesets API used by helpers: **do not edit the benchmark to accommodate the fork** — that would hide the drift we want to measure. Record the exact compiler error, continue, and let the contracts report (Task 11) formally surface the divergence. The Makefile (Task 14) will tolerate a `BUILD_FAILED` fork.

- [ ] **Step 6: Commit**

```bash
git add runners/fork/go.mod runners/fork/go.sum runners/fork/helpers_test.go runners/fork/bench_test.go
git commit -m "feat(fork): add fork runner with replace directive"
```

---

## Task 8: `contracts` Module — Extractor Core (TDD)

**Files:**
- Create: `contracts/go.mod`
- Create: `contracts/internal/extract/extract.go`
- Create: `contracts/internal/extract/extract_test.go`
- Create: `contracts/testdata/fixture/go.mod`
- Create: `contracts/testdata/fixture/fixture.go`

The extractor uses `golang.org/x/tools/go/packages` to load a target module and dump exported analysis symbols. We test it against a tiny fixture module rather than against vacuum directly, so the unit test runs fast and doesn't depend on vacuum's internals.

- [ ] **Step 1: Initialise contracts module**

```bash
mkdir -p contracts/internal/extract contracts/testdata/fixture contracts/cmd/contracts
cd contracts
go mod init vacuum-benchmark/contracts
go get golang.org/x/tools/go/packages
cd ..
```

- [ ] **Step 2: Create the test fixture**

`contracts/testdata/fixture/go.mod`:

```
module fixture

go 1.22
```

`contracts/testdata/fixture/fixture.go`:

```go
// Package fixture is a tiny package used to test the extractor.
package fixture

type Thing struct {
	Name string
	Size int
}

type Doer interface {
	Do() error
}

func ApplyThing(t *Thing) *Thing {
	return t
}

func privateHelper() {}
```

- [ ] **Step 3: Write the failing test**

`contracts/internal/extract/extract_test.go`:

```go
package extract

import (
	"path/filepath"
	"testing"
)

func TestExtractFixture(t *testing.T) {
	fixtureDir, err := filepath.Abs("../../testdata/fixture")
	if err != nil {
		t.Fatal(err)
	}

	got, err := Extract(fixtureDir, []string{"fixture"})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	pkg, ok := got.Packages["fixture"]
	if !ok {
		t.Fatalf("package 'fixture' missing; got keys: %v", keys(got.Packages))
	}

	wantFuncs := map[string]string{
		"ApplyThing": "func(t *Thing) *Thing",
	}
	for name, sig := range wantFuncs {
		found := false
		for _, f := range pkg.Functions {
			if f.Name == name {
				found = true
				if f.Signature != sig {
					t.Errorf("func %s signature = %q, want %q", name, f.Signature, sig)
				}
			}
		}
		if !found {
			t.Errorf("func %s not found; got funcs: %v", name, pkg.Functions)
		}
	}

	for _, f := range pkg.Functions {
		if f.Name == "privateHelper" {
			t.Errorf("unexported privateHelper leaked into output")
		}
	}

	foundThing := false
	for _, typ := range pkg.Types {
		if typ.Name == "Thing" {
			foundThing = true
			if typ.Kind != "struct" {
				t.Errorf("Thing kind = %q, want 'struct'", typ.Kind)
			}
			if len(typ.Fields) != 2 {
				t.Errorf("Thing fields = %d, want 2", len(typ.Fields))
			}
		}
	}
	if !foundThing {
		t.Error("type Thing not found")
	}

	foundDoer := false
	for _, typ := range pkg.Types {
		if typ.Name == "Doer" {
			foundDoer = true
			if typ.Kind != "interface" {
				t.Errorf("Doer kind = %q, want 'interface'", typ.Kind)
			}
		}
	}
	if !foundDoer {
		t.Error("type Doer not found")
	}
}

func keys(m map[string]Package) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
```

- [ ] **Step 4: Verify test fails**

```bash
cd contracts && go test ./internal/extract/... ; cd ..
```

Expected: compile error — `undefined: Extract`, `undefined: Package`, etc.

- [ ] **Step 5: Implement the extractor**

`contracts/internal/extract/extract.go`:

```go
// Package extract dumps exported symbols from a set of Go packages,
// loaded via golang.org/x/tools/go/packages against a target module.
package extract

import (
	"encoding/json"
	"fmt"
	"go/types"
	"sort"

	"golang.org/x/tools/go/packages"
)

type Output struct {
	Version  string             `json:"version"`
	Packages map[string]Package `json:"packages"`
}

type Package struct {
	Path      string     `json:"path"`
	Functions []Function `json:"functions"`
	Types     []TypeInfo `json:"types"`
	Methods   []Method   `json:"methods"`
}

type Function struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
}

type TypeInfo struct {
	Name   string      `json:"name"`
	Kind   string      `json:"kind"` // "struct", "interface", "alias", "named"
	Fields []FieldInfo `json:"fields,omitempty"`
}

type FieldInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Method struct {
	Receiver  string `json:"receiver"`
	Name      string `json:"name"`
	Signature string `json:"signature"`
}

// Extract loads packages in paths (relative to moduleDir) and returns
// their exported symbols.
func Extract(moduleDir string, paths []string) (*Output, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax | packages.NeedImports | packages.NeedDeps,
		Dir: moduleDir,
	}
	pkgs, err := packages.Load(cfg, paths...)
	if err != nil {
		return nil, fmt.Errorf("packages.Load: %w", err)
	}

	out := &Output{Packages: map[string]Package{}}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package %s had errors: %v", pkg.PkgPath, pkg.Errors)
		}
		key := pkg.Name
		if _, taken := out.Packages[key]; taken {
			key = pkg.PkgPath
		}
		out.Packages[key] = extractPackage(pkg)
	}
	return out, nil
}

func extractPackage(pkg *packages.Package) Package {
	p := Package{Path: pkg.PkgPath}
	if pkg.Types == nil {
		return p
	}
	scope := pkg.Types.Scope()
	names := scope.Names()
	sort.Strings(names)

	for _, name := range names {
		obj := scope.Lookup(name)
		if !obj.Exported() {
			continue
		}
		switch o := obj.(type) {
		case *types.Func:
			p.Functions = append(p.Functions, Function{
				Name:      o.Name(),
				Signature: types.TypeString(o.Type(), relativeQualifier(pkg.PkgPath)),
			})
		case *types.TypeName:
			p.Types = append(p.Types, extractType(o, pkg))
			for _, m := range methodsOf(o) {
				p.Methods = append(p.Methods, m)
			}
		}
	}
	return p
}

func extractType(tn *types.TypeName, pkg *packages.Package) TypeInfo {
	info := TypeInfo{Name: tn.Name()}
	underlying := tn.Type().Underlying()

	switch u := underlying.(type) {
	case *types.Struct:
		info.Kind = "struct"
		for i := 0; i < u.NumFields(); i++ {
			f := u.Field(i)
			if !f.Exported() {
				continue
			}
			info.Fields = append(info.Fields, FieldInfo{
				Name: f.Name(),
				Type: types.TypeString(f.Type(), relativeQualifier(pkg.PkgPath)),
			})
		}
	case *types.Interface:
		info.Kind = "interface"
		for i := 0; i < u.NumMethods(); i++ {
			m := u.Method(i)
			if !m.Exported() {
				continue
			}
			info.Fields = append(info.Fields, FieldInfo{
				Name: m.Name(),
				Type: types.TypeString(m.Type(), relativeQualifier(pkg.PkgPath)),
			})
		}
	default:
		info.Kind = "named"
	}
	return info
}

func methodsOf(tn *types.TypeName) []Method {
	var out []Method
	named, ok := tn.Type().(*types.Named)
	if !ok {
		return nil
	}
	for i := 0; i < named.NumMethods(); i++ {
		m := named.Method(i)
		if !m.Exported() {
			continue
		}
		out = append(out, Method{
			Receiver:  tn.Name(),
			Name:      m.Name(),
			Signature: types.TypeString(m.Type(), relativeQualifier(named.Obj().Pkg().Path())),
		})
	}
	return out
}

// relativeQualifier hides the fully-qualified path for types in the same
// package. Cross-package types get their short name.
func relativeQualifier(own string) types.Qualifier {
	return func(other *types.Package) string {
		if other == nil || other.Path() == own {
			return ""
		}
		return other.Name()
	}
}

// MarshalJSON produces deterministic JSON: sorted maps, stable field order.
func (o *Output) MarshalJSON() ([]byte, error) {
	type alias Output
	return json.MarshalIndent((*alias)(o), "", "  ")
}
```

- [ ] **Step 6: Verify test passes**

```bash
cd contracts && go test ./internal/extract/... -v ; cd ..
```

Expected: `PASS` on `TestExtractFixture`.

- [ ] **Step 7: Commit**

```bash
git add contracts/go.mod contracts/go.sum contracts/internal/extract/ contracts/testdata/
git commit -m "feat(contracts): add go/packages-backed symbol extractor"
```

---

## Task 9: `contracts/cmd/contracts` — CLI Wrapper

**Files:**
- Create: `contracts/cmd/contracts/main.go`

- [ ] **Step 1: Write the CLI**

`contracts/cmd/contracts/main.go`:

```go
// Binary contracts dumps exported symbols from the vacuum analysis
// packages (motor, rulesets, model) of whichever vacuum version is
// resolved by the current module.
//
// Usage:
//
//	cd runners/upstream && go run vacuum-benchmark/contracts/cmd/contracts \
//	    -version upstream-v0.26.1 -out ../../results/raw/upstream.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"vacuum-benchmark/contracts/internal/extract"
)

var analysisPackages = []string{
	"github.com/daveshanley/vacuum/motor",
	"github.com/daveshanley/vacuum/rulesets",
	"github.com/daveshanley/vacuum/model",
}

func main() {
	version := flag.String("version", "", "version label, e.g. 'upstream-v0.26.1' or 'fork-<sha>'")
	outPath := flag.String("out", "", "output JSON path (required)")
	moduleDir := flag.String("module-dir", ".", "directory of the target Go module (must contain go.mod)")
	flag.Parse()

	if *version == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: contracts -version <label> -out <path> [-module-dir <dir>]")
		os.Exit(2)
	}

	out, err := extract.Extract(*moduleDir, analysisPackages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "extract failed: %v\n", err)
		os.Exit(1)
	}
	out.Version = *version

	f, err := os.Create(*outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create %s: %v\n", *outPath, err)
		os.Exit(1)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d packages)\n", *outPath, len(out.Packages))
}
```

- [ ] **Step 2: Smoke test against upstream runner**

```bash
mkdir -p results/raw
go -C runners/upstream run vacuum-benchmark/contracts/cmd/contracts \
    -version "upstream-v0.26.1" \
    -out "../../results/raw/upstream.json" \
    -module-dir "."
```

Expected: `wrote ../../results/raw/upstream.json (3 packages)`. Open the file; it should contain non-empty `functions`/`types`/`methods` arrays for the three packages. If a package is missing or errors out, the extractor's error handling surfaces the compile issue — treat as blocker and fix (usually a wrong import path or the target module not being tidied).

- [ ] **Step 3: Smoke test against fork runner**

```bash
go -C runners/fork run vacuum-benchmark/contracts/cmd/contracts \
    -version "fork-HEAD" \
    -out "../../results/raw/fork.json" \
    -module-dir "."
```

Expected: same shape. If the fork fails to load one of the analysis packages, that itself is a drift finding and will surface in the diff — move on.

- [ ] **Step 4: Commit**

```bash
git add contracts/cmd/contracts/main.go
git commit -m "feat(contracts): add cli binary to dump analysis packages"
```

---

## Task 10: `cmd/report` — Bench Output Parser (TDD)

**Files:**
- Create: `cmd/report/go.mod`
- Create: `cmd/report/internal/benchparse/benchparse.go`
- Create: `cmd/report/internal/benchparse/benchparse_test.go`

- [ ] **Step 1: Initialise module**

```bash
mkdir -p cmd/report/internal/benchparse cmd/report/internal/contractdiff cmd/report/internal/render
cd cmd/report
go mod init vacuum-benchmark/cmd/report
cd ../..
```

- [ ] **Step 2: Write the failing test**

`cmd/report/internal/benchparse/benchparse_test.go`:

```go
package benchparse

import (
	"strings"
	"testing"
)

const sample = `
goos: darwin
goarch: arm64
pkg: vacuum-benchmark/runners/upstream
cpu: Apple M2
BenchmarkLint/small/minimal-8     1234    987654 ns/op    45.67 MB/s    34567 B/op    123 allocs/op
BenchmarkLint/small/recommended-8  567    1234567 ns/op   12.34 MB/s   456789 B/op   4567 allocs/op
PASS
ok  	vacuum-benchmark/runners/upstream	2.345s
`

func TestParseSample(t *testing.T) {
	got, err := Parse(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(Parse) = %d, want 2", len(got))
	}

	first := got[0]
	if first.Size != "small" || first.Ruleset != "minimal" {
		t.Errorf("first.Size/Ruleset = %s/%s, want small/minimal", first.Size, first.Ruleset)
	}
	if first.NsPerOp != 987654 {
		t.Errorf("first.NsPerOp = %d, want 987654", first.NsPerOp)
	}
	if first.BytesPerOp != 34567 {
		t.Errorf("first.BytesPerOp = %d, want 34567", first.BytesPerOp)
	}
	if first.AllocsPerOp != 123 {
		t.Errorf("first.AllocsPerOp = %d, want 123", first.AllocsPerOp)
	}
}

func TestParseIgnoresNonBenchLines(t *testing.T) {
	got, err := Parse(strings.NewReader("hello\nworld\nPASS\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("len(Parse) = %d, want 0", len(got))
	}
}
```

- [ ] **Step 3: Verify test fails**

```bash
cd cmd/report && go test ./internal/benchparse/... ; cd ../..
```

Expected: compile error.

- [ ] **Step 4: Implement the parser**

`cmd/report/internal/benchparse/benchparse.go`:

```go
// Package benchparse parses `go test -bench -benchmem` output.
package benchparse

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Result is one row of benchmark output.
type Result struct {
	Name        string // full sub-benchmark name, e.g. "small/minimal"
	Size        string // "small" | "medium" | "large"
	Ruleset     string // "minimal" | "recommended" | ...
	Iterations  int64
	NsPerOp     int64
	BytesPerOp  int64 // from B/op
	AllocsPerOp int64 // from allocs/op
	MBPerSec    float64
}

// Parse reads benchmark output from r and returns one Result per
// `BenchmarkLint/...` line.
func Parse(r io.Reader) ([]Result, error) {
	var out []Result
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "BenchmarkLint/") {
			continue
		}
		res, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", line, err)
		}
		out = append(out, res)
	}
	return out, scanner.Err()
}

func parseLine(line string) (Result, error) {
	// Format: BenchmarkLint/<size>/<ruleset>-<GOMAXPROCS> <iterations> <ns> ns/op [<mbps> MB/s] [<b> B/op] [<allocs> allocs/op]
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return Result{}, fmt.Errorf("too few fields: %d", len(fields))
	}

	name := fields[0]
	// strip the -<GOMAXPROCS> suffix from the name
	if dash := strings.LastIndex(name, "-"); dash > 0 {
		if _, err := strconv.Atoi(name[dash+1:]); err == nil {
			name = name[:dash]
		}
	}
	// name is now "BenchmarkLint/<size>/<ruleset>"
	parts := strings.SplitN(name, "/", 3)
	if len(parts) != 3 {
		return Result{}, fmt.Errorf("unexpected bench name: %s", name)
	}
	size, ruleset := parts[1], parts[2]

	iter, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return Result{}, fmt.Errorf("iterations: %w", err)
	}

	res := Result{
		Name:       size + "/" + ruleset,
		Size:       size,
		Ruleset:    ruleset,
		Iterations: iter,
	}

	// Walk the remaining pairs of (value, unit).
	for i := 2; i+1 < len(fields); i += 2 {
		valStr := fields[i]
		unit := fields[i+1]
		switch unit {
		case "ns/op":
			n, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return Result{}, fmt.Errorf("ns/op: %w", err)
			}
			res.NsPerOp = int64(n)
		case "B/op":
			n, err := strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				return Result{}, fmt.Errorf("B/op: %w", err)
			}
			res.BytesPerOp = n
		case "allocs/op":
			n, err := strconv.ParseInt(valStr, 10, 64)
			if err != nil {
				return Result{}, fmt.Errorf("allocs/op: %w", err)
			}
			res.AllocsPerOp = n
		case "MB/s":
			n, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return Result{}, fmt.Errorf("MB/s: %w", err)
			}
			res.MBPerSec = n
		}
	}
	return res, nil
}
```

- [ ] **Step 5: Verify test passes**

```bash
cd cmd/report && go test ./internal/benchparse/... -v ; cd ../..
```

Expected: `PASS`.

- [ ] **Step 6: Commit**

```bash
git add cmd/report/go.mod cmd/report/internal/benchparse/
git commit -m "feat(report): add go test -bench output parser"
```

---

## Task 11: `cmd/report` — Contract Diff (TDD)

**Files:**
- Create: `cmd/report/internal/contractdiff/contractdiff.go`
- Create: `cmd/report/internal/contractdiff/contractdiff_test.go`

- [ ] **Step 1: Write the failing test**

`cmd/report/internal/contractdiff/contractdiff_test.go`:

```go
package contractdiff

import (
	"testing"
)

func TestDiffIdentical(t *testing.T) {
	a := Snapshot{
		Version: "upstream",
		Packages: map[string]Package{
			"motor": {
				Functions: []Function{{Name: "Apply", Signature: "func() int"}},
			},
		},
	}
	b := a
	b.Version = "fork"
	got := Diff(a, b)
	if got.Verdict != VerdictDropIn {
		t.Errorf("verdict = %s, want %s", got.Verdict, VerdictDropIn)
	}
	if len(got.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(got.Rows))
	}
	if got.Rows[0].Status != StatusMatch {
		t.Errorf("status = %s, want %s", got.Rows[0].Status, StatusMatch)
	}
}

func TestDiffAddedOnly(t *testing.T) {
	a := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{{Name: "Apply", Signature: "func() int"}},
	}}}
	b := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{
			{Name: "Apply", Signature: "func() int"},
			{Name: "ApplyAsync", Signature: "func() int"},
		},
	}}}
	got := Diff(a, b)
	if got.Verdict != VerdictMinorDrift {
		t.Errorf("verdict = %s, want %s", got.Verdict, VerdictMinorDrift)
	}
}

func TestDiffSignatureChanged(t *testing.T) {
	a := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{{Name: "Apply", Signature: "func() int"}},
	}}}
	b := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{{Name: "Apply", Signature: "func() (int, error)"}},
	}}}
	got := Diff(a, b)
	if got.Verdict != VerdictBreaking {
		t.Errorf("verdict = %s, want %s", got.Verdict, VerdictBreaking)
	}
	found := false
	for _, r := range got.Rows {
		if r.Status == StatusSignatureChanged {
			found = true
		}
	}
	if !found {
		t.Error("expected a signature-changed row")
	}
}

func TestDiffRemoved(t *testing.T) {
	a := Snapshot{Packages: map[string]Package{"motor": {
		Functions: []Function{{Name: "Apply", Signature: "func() int"}},
	}}}
	b := Snapshot{Packages: map[string]Package{"motor": {}}}
	got := Diff(a, b)
	if got.Verdict != VerdictBreaking {
		t.Errorf("verdict = %s, want %s", got.Verdict, VerdictBreaking)
	}
}
```

- [ ] **Step 2: Verify test fails**

```bash
cd cmd/report && go test ./internal/contractdiff/... ; cd ../..
```

Expected: compile error.

- [ ] **Step 3: Implement the diff**

`cmd/report/internal/contractdiff/contractdiff.go`:

```go
// Package contractdiff compares two extracted API snapshots and
// classifies the fork's drift.
package contractdiff

import "sort"

type Snapshot struct {
	Version  string             `json:"version"`
	Packages map[string]Package `json:"packages"`
}

type Package struct {
	Path      string     `json:"path"`
	Functions []Function `json:"functions"`
	Types     []TypeInfo `json:"types"`
	Methods   []Method   `json:"methods"`
}

type Function struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
}

type TypeInfo struct {
	Name   string      `json:"name"`
	Kind   string      `json:"kind"`
	Fields []FieldInfo `json:"fields,omitempty"`
}

type FieldInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Method struct {
	Receiver  string `json:"receiver"`
	Name      string `json:"name"`
	Signature string `json:"signature"`
}

type Status string

const (
	StatusMatch             Status = "match"
	StatusSignatureChanged  Status = "signature-changed"
	StatusRemoved           Status = "removed"
	StatusAdded             Status = "added"
)

type Verdict string

const (
	VerdictDropIn     Verdict = "drop-in"
	VerdictMinorDrift Verdict = "minor-drift"
	VerdictBreaking   Verdict = "breaking-drift"
)

type Row struct {
	Package   string
	Kind      string // "func" | "type" | "method"
	Symbol    string // name or receiver.name for methods
	Upstream  string // signature or kind/fields summary
	Fork      string
	Status    Status
}

type Report struct {
	Rows    []Row
	Verdict Verdict
}

// Diff compares upstream vs fork. Order of Rows is deterministic
// (by package, kind, symbol).
func Diff(upstream, fork Snapshot) Report {
	var rows []Row
	pkgs := unionKeys(upstream.Packages, fork.Packages)
	for _, pkg := range pkgs {
		u, hasU := upstream.Packages[pkg]
		f, hasF := fork.Packages[pkg]
		if !hasU {
			rows = append(rows, Row{Package: pkg, Kind: "package", Symbol: pkg, Status: StatusAdded, Fork: "present"})
			continue
		}
		if !hasF {
			rows = append(rows, Row{Package: pkg, Kind: "package", Symbol: pkg, Status: StatusRemoved, Upstream: "present"})
			continue
		}
		rows = append(rows, diffFunctions(pkg, u.Functions, f.Functions)...)
		rows = append(rows, diffTypes(pkg, u.Types, f.Types)...)
		rows = append(rows, diffMethods(pkg, u.Methods, f.Methods)...)
	}
	return Report{Rows: rows, Verdict: computeVerdict(rows)}
}

func diffFunctions(pkg string, u, f []Function) []Row {
	uMap := funcsByName(u)
	fMap := funcsByName(f)
	var rows []Row
	for _, name := range sortedKeys(unionKeys(uMap, fMap)) {
		uFn, hasU := uMap[name]
		fFn, hasF := fMap[name]
		switch {
		case hasU && hasF && uFn.Signature == fFn.Signature:
			rows = append(rows, Row{Package: pkg, Kind: "func", Symbol: name, Upstream: uFn.Signature, Fork: fFn.Signature, Status: StatusMatch})
		case hasU && hasF:
			rows = append(rows, Row{Package: pkg, Kind: "func", Symbol: name, Upstream: uFn.Signature, Fork: fFn.Signature, Status: StatusSignatureChanged})
		case hasU:
			rows = append(rows, Row{Package: pkg, Kind: "func", Symbol: name, Upstream: uFn.Signature, Status: StatusRemoved})
		case hasF:
			rows = append(rows, Row{Package: pkg, Kind: "func", Symbol: name, Fork: fFn.Signature, Status: StatusAdded})
		}
	}
	return rows
}

func diffTypes(pkg string, u, f []TypeInfo) []Row {
	uMap := typesByName(u)
	fMap := typesByName(f)
	var rows []Row
	for _, name := range sortedKeys(unionKeys(uMap, fMap)) {
		uT, hasU := uMap[name]
		fT, hasF := fMap[name]
		switch {
		case hasU && hasF && typeEqual(uT, fT):
			rows = append(rows, Row{Package: pkg, Kind: "type", Symbol: name, Upstream: summarizeType(uT), Fork: summarizeType(fT), Status: StatusMatch})
		case hasU && hasF:
			rows = append(rows, Row{Package: pkg, Kind: "type", Symbol: name, Upstream: summarizeType(uT), Fork: summarizeType(fT), Status: StatusSignatureChanged})
		case hasU:
			rows = append(rows, Row{Package: pkg, Kind: "type", Symbol: name, Upstream: summarizeType(uT), Status: StatusRemoved})
		case hasF:
			rows = append(rows, Row{Package: pkg, Kind: "type", Symbol: name, Fork: summarizeType(fT), Status: StatusAdded})
		}
	}
	return rows
}

func diffMethods(pkg string, u, f []Method) []Row {
	uMap := methodsByKey(u)
	fMap := methodsByKey(f)
	var rows []Row
	for _, key := range sortedKeys(unionKeys(uMap, fMap)) {
		uM, hasU := uMap[key]
		fM, hasF := fMap[key]
		switch {
		case hasU && hasF && uM.Signature == fM.Signature:
			rows = append(rows, Row{Package: pkg, Kind: "method", Symbol: key, Upstream: uM.Signature, Fork: fM.Signature, Status: StatusMatch})
		case hasU && hasF:
			rows = append(rows, Row{Package: pkg, Kind: "method", Symbol: key, Upstream: uM.Signature, Fork: fM.Signature, Status: StatusSignatureChanged})
		case hasU:
			rows = append(rows, Row{Package: pkg, Kind: "method", Symbol: key, Upstream: uM.Signature, Status: StatusRemoved})
		case hasF:
			rows = append(rows, Row{Package: pkg, Kind: "method", Symbol: key, Fork: fM.Signature, Status: StatusAdded})
		}
	}
	return rows
}

func computeVerdict(rows []Row) Verdict {
	hasAdd := false
	for _, r := range rows {
		switch r.Status {
		case StatusSignatureChanged, StatusRemoved:
			return VerdictBreaking
		case StatusAdded:
			hasAdd = true
		}
	}
	if hasAdd {
		return VerdictMinorDrift
	}
	return VerdictDropIn
}

func funcsByName(fs []Function) map[string]Function {
	m := map[string]Function{}
	for _, f := range fs {
		m[f.Name] = f
	}
	return m
}

func typesByName(ts []TypeInfo) map[string]TypeInfo {
	m := map[string]TypeInfo{}
	for _, t := range ts {
		m[t.Name] = t
	}
	return m
}

func methodsByKey(ms []Method) map[string]Method {
	m := map[string]Method{}
	for _, x := range ms {
		m[x.Receiver+"."+x.Name] = x
	}
	return m
}

func typeEqual(a, b TypeInfo) bool {
	if a.Kind != b.Kind || len(a.Fields) != len(b.Fields) {
		return false
	}
	for i := range a.Fields {
		if a.Fields[i] != b.Fields[i] {
			return false
		}
	}
	return true
}

func summarizeType(t TypeInfo) string {
	return t.Kind + "(" + itoa(len(t.Fields)) + " fields)"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func unionKeys[T any](a, b map[string]T) []string {
	seen := map[string]struct{}{}
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	return sortedKeys(mapKeys(seen))
}

func mapKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func sortedKeys(xs []string) []string {
	sort.Strings(xs)
	return xs
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd cmd/report && go test ./internal/contractdiff/... -v ; cd ../..
```

Expected: all four tests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/report/internal/contractdiff/
git commit -m "feat(report): add contract diff with drift verdict"
```

---

## Task 12: `cmd/report` — Markdown Renderer (TDD)

**Files:**
- Create: `cmd/report/internal/render/render.go`
- Create: `cmd/report/internal/render/render_test.go`

- [ ] **Step 1: Write the failing test**

`cmd/report/internal/render/render_test.go`:

```go
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

	for _, want := range []string{"## Ruleset: recommended", "## Ruleset: minimal", "-10%", "Apple M2", "go1.22"} {
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

func TestResultsHandlesBuildFailure(t *testing.T) {
	meta := Meta{UpstreamBuild: "OK", ForkBuild: "BUILD_FAILED: undefined motor.ApplyRulesToRuleSet"}
	got := Results(meta, nil, nil)
	if !strings.Contains(got, "BUILD_FAILED") {
		t.Errorf("expected BUILD_FAILED surface; got:\n%s", got)
	}
}
```

- [ ] **Step 2: Verify test fails**

```bash
cd cmd/report && go test ./internal/render/... ; cd ../..
```

Expected: compile error.

- [ ] **Step 3: Implement the renderer**

`cmd/report/internal/render/render.go`:

```go
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
		b.WriteString("Every analysis-related symbol matches. A `replace github.com/daveshanley/vacuum => github.com/buraksekili/vacuum <sha>` directive is sufficient; no consumer code changes required.\n\n")
	case contractdiff.VerdictMinorDrift:
		b.WriteString("The fork has added symbols but changed nothing upstream callers already use. Still drop-in via `replace`; fork-only features are opt-in.\n\n")
	case contractdiff.VerdictBreaking:
		b.WriteString("The fork has renamed, removed, or re-signed symbols that upstream callers use. Swapping via `replace` requires consumer code changes. See the table below.\n\n")
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
	// Any unexpected rulesets: append after, sorted.
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
	// thousands separator
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
```

- [ ] **Step 4: Wire scenarios dependency into cmd/report go.mod**

```bash
cd cmd/report
go mod edit -require=vacuum-benchmark/internal/scenarios@v0.0.0
go mod tidy
cd ../..
```

- [ ] **Step 5: Verify tests pass**

```bash
cd cmd/report && go test ./internal/render/... -v ; cd ../..
```

Expected: all three tests PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/report/go.mod cmd/report/go.sum cmd/report/internal/render/
git commit -m "feat(report): add markdown renderer for results + contracts"
```

---

## Task 13: `cmd/report` — Orchestrator Main

**Files:**
- Create: `cmd/report/main.go`

- [ ] **Step 1: Write `main.go`**

```go
// Binary report orchestrates the bench + contract workflows and
// writes results/results.md and results/contracts.md.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"vacuum-benchmark/cmd/report/internal/benchparse"
	"vacuum-benchmark/cmd/report/internal/contractdiff"
	"vacuum-benchmark/cmd/report/internal/render"
)

func main() {
	count := flag.Int("count", 5, "go test -count")
	flag.Parse()

	repoRoot, err := findRepoRoot()
	if err != nil {
		die("find repo root: %v", err)
	}

	rawDir := filepath.Join(repoRoot, "results", "raw")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		die("mkdir results/raw: %v", err)
	}

	// 1. Run benchmarks for each runner. Capture output; note build failures.
	upstreamOut, upstreamBuild := runBench(filepath.Join(repoRoot, "runners/upstream"), *count, filepath.Join(rawDir, "upstream.bench.txt"))
	forkOut, forkBuild := runBench(filepath.Join(repoRoot, "runners/fork"), *count, filepath.Join(rawDir, "fork.bench.txt"))

	upstreamResults, _ := benchparse.Parse(bytes.NewReader(upstreamOut))
	forkResults, _ := benchparse.Parse(bytes.NewReader(forkOut))

	// 2. Run contracts extractor for each runner.
	upstreamContract := runContracts(repoRoot, "runners/upstream", "upstream", filepath.Join(rawDir, "upstream.json"))
	forkContract := runContracts(repoRoot, "runners/fork", "fork", filepath.Join(rawDir, "fork.json"))

	// 3. Diff.
	diffReport := contractdiff.Diff(upstreamContract, forkContract)

	// 4. Render.
	meta := render.Meta{
		GoVersion:     runtime.Version(),
		CPU:           readCPU(),
		Date:          time.Now().Format("2006-01-02"),
		Count:         *count,
		Upstream:      upstreamContract.Version,
		Fork:          forkContract.Version,
		UpstreamBuild: upstreamBuild,
		ForkBuild:     forkBuild,
	}
	resultsMD := render.Results(meta, upstreamResults, forkResults)
	contractsMD := render.Contracts(meta, diffReport)

	writeFile(filepath.Join(repoRoot, "results/results.md"), resultsMD)
	writeFile(filepath.Join(repoRoot, "results/contracts.md"), contractsMD)

	fmt.Println("wrote results/results.md and results/contracts.md")
}

// runBench runs `go test -bench=BenchmarkLint -benchmem -run=^$ -count=N ./...`
// in the given runner directory. Returns the captured output and a build status
// string ("OK" or "BUILD_FAILED: <trimmed error>").
func runBench(runnerDir string, count int, rawPath string) ([]byte, string) {
	cmd := exec.Command("go", "test", "-bench=BenchmarkLint", "-benchmem", "-run=^$",
		fmt.Sprintf("-count=%d", count), "./...")
	cmd.Dir = runnerDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	// Always persist raw stdout+stderr for debugging.
	_ = os.WriteFile(rawPath, append(stdout.Bytes(), stderr.Bytes()...), 0o644)

	if err != nil {
		combined := strings.TrimSpace(stderr.String() + "\n" + stdout.String())
		if combined == "" {
			combined = err.Error()
		}
		// Trim to something presentable in the report header.
		if len(combined) > 200 {
			combined = combined[:200] + "..."
		}
		return stdout.Bytes(), "BUILD_FAILED: " + combined
	}
	return stdout.Bytes(), "OK"
}

// runContracts invokes the contracts CLI with cwd=runnerDir. Returns an empty
// snapshot if the invocation fails (the diff will flag the missing packages).
func runContracts(repoRoot, runnerRel, versionLabel, outPath string) contractdiff.Snapshot {
	cmd := exec.Command("go", "run", "vacuum-benchmark/contracts/cmd/contracts",
		"-version", versionLabel,
		"-out", outPath,
		"-module-dir", ".")
	cmd.Dir = filepath.Join(repoRoot, runnerRel)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "contracts for %s failed: %v\n", runnerRel, err)
		return contractdiff.Snapshot{Version: versionLabel + " (extract failed)", Packages: map[string]contractdiff.Package{}}
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		die("read %s: %v", outPath, err)
	}
	var snap contractdiff.Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		die("unmarshal %s: %v", outPath, err)
	}
	return snap
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.work not found")
		}
		dir = parent
	}
}

func readCPU() string {
	out, err := exec.Command("uname", "-m").Output()
	if err != nil {
		return runtime.GOARCH
	}
	return strings.TrimSpace(string(out)) + " " + runtime.GOOS
}

func writeFile(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		die("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		die("write %s: %v", path, err)
	}
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
```

- [ ] **Step 2: Compile check**

```bash
cd cmd/report && go build ./... ; cd ../..
```

Expected: clean build. If `contractdiff.Snapshot` types can't unmarshal the extractor's JSON because field names differ, reconcile — the two types in `contracts/internal/extract` and `cmd/report/internal/contractdiff` intentionally mirror each other; JSON tags must match. Check side-by-side if tests pass but real invocation fails at unmarshal time.

- [ ] **Step 3: End-to-end dry run**

```bash
./scripts/fetch.sh  # skip if already run
go -C cmd/report run . -count=1
```

Expected:
- `results/raw/upstream.bench.txt`, `fork.bench.txt` exist (raw bench output)
- `results/raw/upstream.json`, `fork.json` exist (extractor output)
- `results/results.md`, `results/contracts.md` exist
- Open `results/results.md`: 3 tables (one per ruleset × 3 sizes? actually one per ruleset, 3 rows each) plus header with metadata
- Open `results/contracts.md`: verdict + parity table

If the fork's BenchmarkLint failed to build, `results/results.md` should still render — upstream columns filled, fork columns dashes — with the fork's error surfaced in the header.

- [ ] **Step 4: Commit**

```bash
git add cmd/report/main.go
git commit -m "feat(report): add orchestrator that runs benches+contracts and writes markdown"
```

---

## Task 14: Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Write Makefile**

```makefile
# vacuum-benchmark

COUNT ?= 5

.PHONY: all fetch bench contracts clean tidy

all: fetch bench contracts

fetch:
	./scripts/fetch.sh

bench: fetch
	@mkdir -p results/raw
	cd cmd/report && go run . -count=$(COUNT)

contracts:
	@mkdir -p results/raw
	cd cmd/report && go run . -count=0 >/dev/null 2>&1 || true
	@echo "contracts.md written alongside bench output. For contracts-only, re-run 'make bench' with any COUNT (contracts are always regenerated)."

tidy:
	cd internal/scenarios && go mod tidy
	cd runners/upstream    && go mod tidy
	cd runners/fork        && go mod tidy
	cd contracts           && go mod tidy
	cd cmd/report          && go mod tidy

clean:
	rm -rf results/
```

Note: the orchestrator writes both `results.md` and `contracts.md` in one pass. A dedicated `make contracts` that skips benchmarks would require splitting the orchestrator; not worth the complexity — `make bench COUNT=1` is the fast path. The `contracts` target exists as documentation that the two reports are produced together.

- [ ] **Step 2: Smoke test**

```bash
make tidy
make bench COUNT=1
```

Expected: `results/results.md` and `results/contracts.md` regenerated. No errors (fork may still BUILD_FAILED; that's reflected in the report, not a Makefile failure).

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "feat: add Makefile with fetch/bench/contracts/clean targets"
```

---

## Task 15: README Usage Documentation

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Extend README with usage + interpretation**

Overwrite `README.md`:

```markdown
# vacuum-benchmark

A reproducible benchmark comparing two Go OpenAPI linter implementations:

- **Upstream:** [`github.com/daveshanley/vacuum`](https://github.com/daveshanley/vacuum) pinned to `v0.26.1`
- **Fork:** [`github.com/buraksekili/vacuum`](https://github.com/buraksekili/vacuum) at a pinned commit (see `runners/fork/go.mod`)

## Quick Start

```bash
make fetch      # download real-world OpenAPI specs into testdata/openapis/
make all        # run benchmarks + API-contract diff
```

Outputs: `results/results.md`, `results/contracts.md`.

## What gets measured

**Performance (`results/results.md`)** — for each of 3 spec sizes × 5 rulesets = 15 scenarios, `BenchmarkLint` measures end-to-end `motor.ApplyRulesToRuleSet`: ns/op, B/op, allocs/op, MB/s. The orchestrator runs `go test -bench -benchmem` against each runner and produces one comparison table per ruleset, with a delta column.

**API contract parity (`results/contracts.md`)** — exported symbols in the analysis-related packages (`motor`, `rulesets`, `model`) are extracted from each runner via `golang.org/x/tools/go/packages`, then diffed. Output is a parity table plus a one-line verdict:

- **drop-in** — symbol-for-symbol match; `replace` is sufficient
- **minor-drift** — fork added symbols; existing callers unaffected
- **breaking-drift** — fork renamed/removed/re-signed symbols; consumers must update

## Layout

- `testdata/openapis/` — fetched OpenAPI specs (gitignored; run `make fetch`)
- `testdata/rulesets/` — five committed rulesets
- `internal/scenarios/` — shared scenario matrix (no vacuum import)
- `runners/upstream/`, `runners/fork/` — per-version modules; each pins its own vacuum
- `contracts/` — API-surface extractor
- `cmd/report/` — orchestrator + markdown renderer
- `scripts/fetch.sh` — SHA-pinned spec fetcher

## Updating pinned versions

- **Upstream vacuum:** edit `runners/upstream/go.mod`, bump the `require` line, `cd runners/upstream && go mod tidy`.
- **Fork vacuum:** edit `runners/fork/go.mod`, change the commit SHA on the `replace` line, `cd runners/fork && go mod tidy`.
- **OpenAPI specs:** edit the SHA constants at the top of `scripts/fetch.sh`, re-run `make fetch`.

## Known limitations

- Benchmarks are sensitive to laptop thermal state and background processes. Re-run with higher `COUNT` to reduce noise: `make bench COUNT=10`.
- If the fork has diverged on `motor.ApplyRulesToRuleSet` or `RuleSetExecution`, the fork runner won't compile. The report surfaces this as `BUILD_FAILED` in the header and the contract diff will show the divergent symbols — this is the intended failure mode.
- Only YAML specs are benchmarked; JSON parsing cost is not measured.
- Fork staleness (commits-behind-upstream) is out of scope — this project measures runtime behaviour and API surface, not git history.
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: expand readme with usage, interpretation, and limitations"
```

---

## Task 16: Final End-to-End Verification

**Files:** none (verification only)

- [ ] **Step 1: Clean slate**

```bash
make clean
rm -rf testdata/openapis/*.yaml
```

- [ ] **Step 2: Full run**

```bash
make all COUNT=2
```

Expected:
- `make fetch` downloads three `.yaml` files
- `make bench` runs both runners; fork may `BUILD_FAILED` — acceptable
- `results/results.md` exists and is non-empty, tables render
- `results/contracts.md` exists, has a verdict line

- [ ] **Step 3: Open and eyeball the reports**

```bash
cat results/results.md
cat results/contracts.md
```

Confirm: header records Go version, CPU, date, count, and per-runner build status. Delta column shows `+x.x%` or `-x.x%` for scenarios where both runners succeeded.

- [ ] **Step 4: Final commit (if anything changed during verification)**

```bash
git status
```

If clean, done. If any files changed (e.g., generated artefacts leaked past `.gitignore`), add entries to `.gitignore` and commit:

```bash
git add .gitignore
git commit -m "chore: ignore additional generated artefacts"
```

- [ ] **Step 5: Log summary**

Echo a short summary to stdout of what the benchmark produced — include a few representative deltas and the drift verdict. No commit needed; this is a human-facing sanity check.

---

## Self-Review

**Spec coverage:**

- ✅ Repo layout (Task 1, 4, 5, 7, 8, 9, 10, 13)
- ✅ Go workspace (Task 1)
- ✅ `testdata/rulesets/` with five rulesets (Task 2)
- ✅ `testdata/openapis/` fetched via SHA-pinned script (Task 3)
- ✅ `internal/scenarios` pure-data module (Task 4)
- ✅ `runners/upstream` pinning v0.26.1 (Task 5, 6)
- ✅ `runners/fork` with `replace` directive (Task 7)
- ✅ BenchmarkLint measuring end-to-end ApplyRulesToRuleSet with `ReportAllocs` + `SetBytes` (Task 6)
- ✅ Contract extractor on motor/rulesets/model (Task 8, 9)
- ✅ Deterministic JSON output (Task 8)
- ✅ Diff classifying drop-in / minor / breaking (Task 11)
- ✅ Markdown rendering with delta column and build-status surface (Task 12)
- ✅ Orchestrator shelling out to `go test -bench` (Task 13)
- ✅ Makefile with fetch, bench, contracts, clean, tidy (Task 14)
- ✅ README with usage and limitations (Task 15)
- ✅ End-to-end reproducibility verification (Task 16)

Out of scope per spec (intentionally absent): fork staleness measurement, JSON-format specs, parse-only vs rules-only split, correctness diffing of rule results.

**Type consistency:** `contractdiff.Snapshot`/`Package`/etc. mirror `extract.Output`/`Package` with identical JSON tags (verified in Task 13 Step 2). `render.Meta` fields align with what Task 13 populates (GoVersion, CPU, Date, Count, Upstream, Fork, UpstreamBuild, ForkBuild). Benchmark name format (`BenchmarkLint/<size>/<ruleset>`) is consistent across Task 6 (producer) and Task 10 (parser).

**Placeholder scan:** only placeholders are the three GitHub SHAs in Task 3 and the fork SHA in Task 7. Each is resolved by a concrete `gh api` command in the same task, not left for the engineer to guess.
