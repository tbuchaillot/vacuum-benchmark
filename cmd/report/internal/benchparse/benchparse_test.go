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
