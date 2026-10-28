// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestComplexityCountsEachBranch(t *testing.T) {
	// Arrange
	tests := []struct {
		name string
		body string
		want int
	}{
		{"straight line", "x := 1\n_ = x", 1},
		{"if and else", "if a { } else { }", 2},
		{"for and range", "for i := 0; i < 3; i++ { }\nfor range s { }", 3},
		{"switch cases, default not counted", "switch a { case 1: case 2, 3: default: }", 3},
		{"select cases", "select { case <-c: default: }", 2},
		{"and, or", "if a && b || c { }", 4},
		{"closure counts toward the function", "f := func() { if a { } }\nf()", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			src := "package p\nvar a, b, c bool\nvar s []int\nvar ch chan int\nfunc F() {\n" + strings.ReplaceAll(tt.body, "<-c", "<-ch") + "\n}\n"

			// Act
			funcs, err := ParseFuncs("p.go", []byte(src))
			if err != nil {
				t.Fatal(err)
			}

			// Assert
			if len(funcs) != 1 || funcs[0].Complexity != tt.want {
				t.Errorf("complexity = %v, want %d", funcs, tt.want)
			}
		})
	}
}

func TestFuncNameIncludesTheReceiver(t *testing.T) {
	// Arrange
	src := "package p\ntype T struct{}\nfunc (t *T) Load() {}\nfunc (T) Name() string { return \"\" }\nfunc Run() {}\n"
	want := []string{"(*T).Load", "(T).Name", "Run"}

	// Act
	funcs, err := ParseFuncs("p.go", []byte(src))
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	var got []string
	for _, f := range funcs {
		got = append(got, f.Name)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("names (-want +got):\n%s", diff)
	}
}

func TestGroupSortsFilesIntoColumns(t *testing.T) {
	// Arrange
	tests := map[string]string{
		"internal/rules/check.go":      groupCode,
		"cmd/ci-parity/main.go":        groupCode,
		"internal/rules/check_test.go": groupTests,
		"test/perf/perf_test.go":       groupTests,
		"site/scripts/sync-docs.mjs":   groupSite,
	}

	for path, want := range tests {
		// Act
		got := Group(path)

		// Assert
		if got != want {
			t.Errorf("Group(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestParseSCCReadsFilesSortedByPath(t *testing.T) {
	// Arrange
	in := `[{"Name":"Go","Files":[{"Location":"./b.go","Code":5},{"Location":"a.go","Code":3}]},` +
		`{"Name":"CSS","Files":[{"Location":"site/x.css","Code":2}]}]`
	want := []File{{"a.go", 3}, {"b.go", 5}, {"site/x.css", 2}}

	// Act
	got, err := ParseSCC(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("files (-want +got):\n%s", diff)
	}
}

func TestRenderTotalsEachColumn(t *testing.T) {
	// Arrange
	files := []File{{"internal/a.go", 1200}, {"internal/a_test.go", 30}, {"site/x.mjs", 10}}
	funcs := []Func{
		{"F", "internal/a.go", 3, 4},
		{"G", "internal/a.go", 9, 2},
		{"TestF", "internal/a_test.go", 5, 1},
	}

	// Act
	got := Render(files, funcs)

	// Assert
	for _, want := range []string{
		"| Metric | Go code | Go tests | Docs site | Total |",
		"| Lines of code | 1,200 | 30 | 10 | 1,240 |",
		"| Functions | 2 | 1 | – | 3 |",
		"| Average complexity | 3.0 | 1.0 | – | 2.3 |",
		"| `F` | `internal/a.go:3` | 4 |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Render() has no line %q in:\n%s", want, got)
		}
	}
}

func TestReplaceBlockSwapsOnlyTheMarkedBlock(t *testing.T) {
	// Arrange
	readme := "# Title\n\n<!-- METRICS:START -->\nold\n<!-- METRICS:END -->\n\nAfter.\n"
	want := "# Title\n\n<!-- METRICS:START -->\n\n| new |\n\n<!-- METRICS:END -->\n\nAfter.\n"

	// Act
	got, err := ReplaceBlock(readme, "| new |\n")
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if got != want {
		t.Errorf("ReplaceBlock() =\n%q\nwant\n%q", got, want)
	}
}

func TestReplaceBlockNeedsTheMarkers(t *testing.T) {
	// Act
	_, err := ReplaceBlock("# No block\n", "x")

	// Assert
	if err == nil {
		t.Error("ReplaceBlock without markers: want an error")
	}
}
