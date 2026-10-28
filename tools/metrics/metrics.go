// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"slices"
	"strings"
)

// Groups are the README table's columns, in order.
const (
	groupCode  = "Go code"
	groupTests = "Go tests"
	groupSite  = "Docs site"
)

var groups = []string{groupCode, groupTests, groupSite}

// File is one counted file: its repo-relative path and lines of code.
type File struct {
	Path string
	Code int
}

// Func is one Go function or method and its cyclomatic complexity.
type Func struct {
	Name       string
	Path       string
	Line       int
	Complexity int
}

// sccLanguage is one language entry in `scc --format json --by-file` output.
type sccLanguage struct {
	Files []struct {
		Location string `json:"Location"`
		Code     int    `json:"Code"`
	} `json:"Files"`
}

// ParseSCC reads scc's JSON output into files, sorted by path.
func ParseSCC(r io.Reader) ([]File, error) {
	var langs []sccLanguage
	if err := json.NewDecoder(r).Decode(&langs); err != nil {
		return nil, fmt.Errorf("read scc output: %w", err)
	}
	var files []File
	for _, l := range langs {
		for _, f := range l.Files {
			files = append(files, File{Path: strings.TrimPrefix(f.Location, "./"), Code: f.Code})
		}
	}
	slices.SortFunc(files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return files, nil
}

// Group names the column a file belongs to: the docs site, Go tests (test
// files and test/), or Go code.
func Group(path string) string {
	switch {
	case strings.HasPrefix(path, "site/"):
		return groupSite
	case strings.HasSuffix(path, "_test.go") || strings.HasPrefix(path, "test/"):
		return groupTests
	default:
		return groupCode
	}
}

// Complexity returns the cyclomatic complexity of fn: 1, plus one for each if,
// for, range, non-default case or select case, && and ||. Closures count
// toward the function that contains them.
func Complexity(fn *ast.FuncDecl) int {
	n := 1
	ast.Inspect(fn, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			n++
		case *ast.CaseClause:
			if x.List != nil {
				n++
			}
		case *ast.CommClause:
			if x.Comm != nil {
				n++
			}
		case *ast.BinaryExpr:
			if x.Op == token.LAND || x.Op == token.LOR {
				n++
			}
		}
		return true
	})
	return n
}

// FuncName is fn's name, with its receiver type for a method: `(*Mise).Load`.
func FuncName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	var recv strings.Builder
	switch t := fn.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		recv.WriteString("*")
		if id, ok := t.X.(*ast.Ident); ok {
			recv.WriteString(id.Name)
		}
	case *ast.Ident:
		recv.WriteString(t.Name)
	}
	return "(" + recv.String() + ")." + fn.Name.Name
}

// ParseFuncs returns the functions declared in the Go source src, read from path.
func ParseFuncs(path string, src []byte) ([]Func, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	var funcs []Func
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		funcs = append(funcs, Func{
			Name: FuncName(fn), Path: path, Line: fset.Position(fn.Pos()).Line, Complexity: Complexity(fn),
		})
	}
	return funcs, nil
}

// column is one column's totals.
type column struct {
	files, code, funcs, sumComplexity, maxComplexity int
}

// Render writes the README's metrics block: a table of files, lines of code,
// functions and complexity per column, and the 10 most complex functions. It
// has no dates, so the same code always renders the same block.
func Render(files []File, funcs []Func) string {
	cols := map[string]*column{}
	for _, g := range groups {
		cols[g] = &column{}
	}
	total := &column{}
	for _, f := range files {
		c := cols[Group(f.Path)]
		c.files++
		c.code += f.Code
		total.files++
		total.code += f.Code
	}
	for _, fn := range funcs {
		for _, c := range []*column{cols[Group(fn.Path)], total} {
			c.funcs++
			c.sumComplexity += fn.Complexity
			c.maxComplexity = max(c.maxComplexity, fn.Complexity)
		}
	}

	var b strings.Builder
	b.WriteString("| Metric | " + strings.Join(groups, " | ") + " | Total |\n")
	b.WriteString("| --- |" + strings.Repeat(" --- |", len(groups)+1) + "\n")
	all := append(slices.Clone(groups), "Total")
	row := func(label string, value func(*column) string) {
		b.WriteString("| " + label)
		for _, g := range all {
			c := total
			if g != "Total" {
				c = cols[g]
			}
			b.WriteString(" | " + value(c))
		}
		b.WriteString(" |\n")
	}
	row("Files", func(c *column) string { return thousands(c.files) })
	row("Lines of code", func(c *column) string { return thousands(c.code) })
	row("Functions", func(c *column) string { return orDash(c.funcs, thousands(c.funcs)) })
	row("Average complexity", func(c *column) string {
		return orDash(c.funcs, fmt.Sprintf("%.1f", float64(c.sumComplexity)/float64(max(c.funcs, 1))))
	})
	row("Max complexity", func(c *column) string { return orDash(c.funcs, fmt.Sprint(c.maxComplexity)) })

	top := slices.Clone(funcs)
	slices.SortFunc(top, func(a, b Func) int {
		if a.Complexity != b.Complexity {
			return b.Complexity - a.Complexity
		}
		if a.Path != b.Path {
			return strings.Compare(a.Path, b.Path)
		}
		return a.Line - b.Line
	})
	b.WriteString("\n<details>\n<summary>The 10 most complex functions</summary>\n\n")
	b.WriteString("| Function | Where | Complexity |\n| --- | --- | --- |\n")
	for _, fn := range top[:min(10, len(top))] {
		fmt.Fprintf(&b, "| `%s` | `%s:%d` | %d |\n", fn.Name, fn.Path, fn.Line, fn.Complexity)
	}
	b.WriteString("\n</details>\n")
	return b.String()
}

// orDash returns s, or a dash when a column has no functions to measure.
func orDash(n int, s string) string {
	if n == 0 {
		return "–"
	}
	return s
}

// thousands formats n with comma separators: 6419 as 6,419.
func thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// Markers delimit the generated block in the README.
const (
	startMarker = "<!-- METRICS:START -->"
	endMarker   = "<!-- METRICS:END -->"
)

// ReplaceBlock returns readme with the text between the markers replaced by
// block, surrounded by blank lines.
func ReplaceBlock(readme, block string) (string, error) {
	start := strings.Index(readme, startMarker)
	end := strings.Index(readme, endMarker)
	if start < 0 || end < start {
		return "", fmt.Errorf("README has no %s ... %s block", startMarker, endMarker)
	}
	return readme[:start+len(startMarker)] + "\n\n" + block + "\n" + readme[end:], nil
}
