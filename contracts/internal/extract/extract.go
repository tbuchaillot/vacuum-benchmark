// Package extract dumps exported symbols from a set of Go packages,
// loaded via golang.org/x/tools/go/packages against a target module.
//
// Package keys in the Output are the SHORT name (e.g., "motor") rather
// than the full import path. This lets the diff compare the same
// logical package across two vacuum forks whose module-paths differ
// (github.com/daveshanley/vacuum/motor vs github.com/buraksekili/vacuum/motor
// both map to key "motor"). The full import path is preserved in
// Package.Path.
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
// their exported symbols keyed by short package name.
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

	// Aliases (type X = Y) must be distinguished from named types (type X Y)
	// so the diff can surface an alias↔named conversion as drift.
	if tn.IsAlias() {
		info.Kind = "alias"
		info.Fields = []FieldInfo{{
			Name: "",
			Type: types.TypeString(tn.Type(), relativeQualifier(pkg.PkgPath)),
		}}
		return info
	}

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
