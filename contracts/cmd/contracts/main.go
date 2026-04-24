// Binary contracts dumps exported symbols from a list of Go packages
// (scoped to a target module) into a JSON file. Used by the vacuum
// benchmark's orchestrator to snapshot each runner's analysis API.
//
// Usage:
//
//	cd runners/upstream && go run vacuum-benchmark/contracts/cmd/contracts \
//	    -version upstream-v0.26.1 \
//	    -packages github.com/daveshanley/vacuum/motor,github.com/daveshanley/vacuum/rulesets,github.com/daveshanley/vacuum/model \
//	    -out ../../results/raw/upstream.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"vacuum-benchmark/contracts/internal/extract"
)

func main() {
	version := flag.String("version", "", "version label, e.g. 'upstream-v0.26.1' or 'fork-<sha>'")
	outPath := flag.String("out", "", "output JSON path (required)")
	moduleDir := flag.String("module-dir", ".", "directory of the target Go module (must contain go.mod)")
	packagesCSV := flag.String("packages", "", "comma-separated list of import paths to extract (required)")
	flag.Parse()

	if *version == "" || *outPath == "" || *packagesCSV == "" {
		fmt.Fprintln(os.Stderr, "usage: contracts -version <label> -out <path> -packages <csv> [-module-dir <dir>]")
		flag.PrintDefaults()
		os.Exit(2)
	}

	paths := splitCSV(*packagesCSV)
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "no packages after splitting -packages csv")
		os.Exit(2)
	}

	out, err := extract.Extract(*moduleDir, paths)
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
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		f.Close()
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		os.Exit(1)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "close %s: %v\n", *outPath, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d packages)\n", *outPath, len(out.Packages))
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
