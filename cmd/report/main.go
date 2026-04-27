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

// analysisPackages are the vacuum packages whose public API we diff,
// parameterised by the runner's actual module path.
func analysisPackages(modulePrefix string) string {
	return modulePrefix + "/motor," + modulePrefix + "/rulesets," + modulePrefix + "/model"
}

const (
	upstreamModule = "github.com/daveshanley/vacuum"
	forkModule     = "github.com/buraksekili/vacuum"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// run holds the orchestrator body. It's separated from main so deferred
// cleanups (temp binary, temp dir) actually execute on error paths —
// os.Exit in main would skip them.
func run() error {
	count := flag.Int("count", 5, "go test -count")
	flag.Parse()

	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	rawDir := filepath.Join(repoRoot, "results", "raw")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		return fmt.Errorf("mkdir results/raw: %w", err)
	}

	// 1. Build the contracts binary once (from the workspace) so each
	//    invocation against a runner can use it without requiring the
	//    runner to be a workspace member.
	contractsBin, cleanupBin, err := buildContractsBinary(repoRoot)
	if err != nil {
		return fmt.Errorf("build contracts binary: %w", err)
	}
	defer cleanupBin()

	// 2. Run benchmarks for each runner. Capture output; note build failures.
	upstreamOut, upstreamBuild := runBench(filepath.Join(repoRoot, "runners/upstream"), *count, filepath.Join(rawDir, "upstream.bench.txt"))
	forkOut, forkBuild := runBench(filepath.Join(repoRoot, "runners/fork"), *count, filepath.Join(rawDir, "fork.bench.txt"))

	upstreamResults, _ := benchparse.Parse(bytes.NewReader(upstreamOut))
	forkResults, _ := benchparse.Parse(bytes.NewReader(forkOut))

	// 3. Run contracts extractor for each runner, each with its own -packages.
	upstreamContract := runContracts(contractsBin, filepath.Join(repoRoot, "runners/upstream"), "upstream-"+upstreamModule, analysisPackages(upstreamModule), filepath.Join(rawDir, "upstream.json"))
	forkContract := runContracts(contractsBin, filepath.Join(repoRoot, "runners/fork"), "fork-"+forkModule, analysisPackages(forkModule), filepath.Join(rawDir, "fork.json"))

	// 4. Diff.
	diffReport := contractdiff.Diff(upstreamContract, forkContract)

	// 5. Render.
	meta := render.Meta{
		GoVersion:     runtime.Version(),
		CPU:           cpuFromBench(upstreamOut, forkOut),
		Date:          time.Now().Format("2006-01-02"),
		Count:         *count,
		Upstream:      upstreamContract.Version,
		Fork:          forkContract.Version,
		UpstreamBuild: upstreamBuild,
		ForkBuild:     forkBuild,
	}
	resultsMD := render.Results(meta, upstreamResults, forkResults)
	contractsMD := render.Contracts(meta, diffReport)

	if err := writeFile(filepath.Join(repoRoot, "results/results.md"), resultsMD); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(repoRoot, "results/contracts.md"), contractsMD); err != nil {
		return err
	}

	fmt.Println("wrote results/results.md and results/contracts.md")
	return nil
}

// buildContractsBinary compiles the contracts CLI into a fresh temp
// directory (0o700) and returns the executable path alongside a cleanup
// func the caller defers. Using MkdirTemp + a fixed filename avoids the
// TOCTOU window that CreateTemp+Remove+go-build would open in /tmp.
func buildContractsBinary(repoRoot string) (string, func(), error) {
	dir, err := os.MkdirTemp("", "vacuum-bench-contracts-")
	if err != nil {
		return "", nil, fmt.Errorf("mkdir temp: %w", err)
	}
	cleanup := func() { os.RemoveAll(dir) }

	path := filepath.Join(dir, "contracts")
	cmd := exec.Command("go", "build", "-o", path, "./cmd/contracts")
	cmd.Dir = filepath.Join(repoRoot, "contracts")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("go build: %w\n%s", err, stderr.String())
	}
	return path, cleanup, nil
}

// runBench runs `go test -bench=BenchmarkLint -benchmem -run=^$ -count=N ./...`
// in the given runner directory with GOWORK=off (runners are intentionally
// outside the workspace to preserve pinned-dep isolation). Returns the
// captured stdout and a build status ("OK" or "BUILD_FAILED: <trimmed>").
func runBench(runnerDir string, count int, rawPath string) ([]byte, string) {
	cmd := exec.Command("go", "test", "-bench=BenchmarkLint", "-benchmem", "-run=^$",
		fmt.Sprintf("-count=%d", count), "./...")
	cmd.Dir = runnerDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	_ = os.WriteFile(rawPath, append(stdout.Bytes(), stderr.Bytes()...), 0o644)

	if err != nil {
		combined := strings.TrimSpace(stderr.String() + "\n" + stdout.String())
		if combined == "" {
			combined = err.Error()
		}
		if len(combined) > 200 {
			combined = combined[:200] + "..."
		}
		return stdout.Bytes(), "BUILD_FAILED: " + combined
	}
	return stdout.Bytes(), "OK"
}

// runContracts invokes the contracts binary with cwd=runnerDir.
// Returns an empty snapshot if invocation fails.
func runContracts(bin, runnerDir, versionLabel, packagesCSV, outPath string) contractdiff.Snapshot {
	cmd := exec.Command(bin,
		"-version", versionLabel,
		"-out", outPath,
		"-packages", packagesCSV,
		"-module-dir", ".")
	cmd.Dir = runnerDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "contracts for %s failed: %v\n", runnerDir, err)
		return contractdiff.Snapshot{Version: versionLabel + " (extract failed)", Packages: map[string]contractdiff.Package{}}
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "contracts: read %s: %v\n", outPath, err)
		return contractdiff.Snapshot{Version: versionLabel + " (read failed)", Packages: map[string]contractdiff.Package{}}
	}
	var snap contractdiff.Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		fmt.Fprintf(os.Stderr, "contracts: unmarshal %s: %v\n", outPath, err)
		return contractdiff.Snapshot{Version: versionLabel + " (unmarshal failed)", Packages: map[string]contractdiff.Package{}}
	}
	return snap
}

// findRepoRoot walks up from cwd until it finds go.work.
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

// cpuFromBench extracts the CPU model from either runner's bench output.
// `go test -bench` emits a `cpu: <model>` header line (e.g. `cpu: Apple M5 Pro`);
// that is the authoritative source here since we're running on the same
// machine as the benchmarks. Falls back to runtime identifiers if both
// runners failed to produce output (e.g., both BUILD_FAILED).
func cpuFromBench(outputs ...[]byte) string {
	for _, out := range outputs {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if cpu, ok := strings.CutPrefix(line, "cpu: "); ok {
				return cpu
			}
		}
	}
	return runtime.GOARCH + " " + runtime.GOOS
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
