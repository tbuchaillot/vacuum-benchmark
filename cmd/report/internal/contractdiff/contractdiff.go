// Package contractdiff compares two extracted API snapshots and
// classifies the fork's drift.
package contractdiff

import (
	"sort"
	"strconv"
)

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
	StatusMatch            Status = "match"
	StatusSignatureChanged Status = "signature-changed"
	StatusRemoved          Status = "removed"
	StatusAdded            Status = "added"
)

type Verdict string

const (
	VerdictDropIn     Verdict = "drop-in"
	VerdictMinorDrift Verdict = "minor-drift"
	VerdictBreaking   Verdict = "breaking-drift"
)

type Row struct {
	Package  string
	Kind     string // "func" | "type" | "method" | "package"
	Symbol   string // name or receiver.name for methods
	Upstream string // signature or kind summary
	Fork     string
	Status   Status
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
	return t.Kind + "(" + strconv.Itoa(len(t.Fields)) + " fields)"
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
