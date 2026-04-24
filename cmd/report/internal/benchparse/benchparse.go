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
