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

	// 1. Build the contracts binary once (from the workspace) so each
	//    invocation against a runner can use it without requiring the
	//    runner to be a workspace member.
	contractsBin, err := buildContractsBinary(repoRoot)
	if err != nil {
		die("build contracts binary: %v", err)
	}
	defer os.Remove(contractsBin)

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

// buildContractsBinary compiles the contracts CLI to a temp path.
// The caller is responsible for os.Remove-ing the returned path.
func buildContractsBinary(repoRoot string) (string, error) {
	tmpFile, err := os.CreateTemp("", "vacuum-bench-contracts-*")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	path := tmpFile.Name()
	tmpFile.Close()
	os.Remove(path)

	cmd := exec.Command("go", "build", "-o", path, "./cmd/contracts")
	cmd.Dir = filepath.Join(repoRoot, "contracts")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go build: %w\n%s", err, stderr.String())
	}
	return path, nil
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
		die("read %s: %v", outPath, err)
	}
	var snap contractdiff.Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		die("unmarshal %s: %v", outPath, err)
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
