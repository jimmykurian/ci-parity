package domain

import (
	"fmt"
	"sort"
)

// Baseline is the per-file count of tolerated inline-tool findings.
type Baseline struct {
	Workflows map[string]int
}

// Row compares one scanned file's current inline-tool count to its baseline.
type Row struct {
	File     string
	Baseline int
	Current  int
}

// Locked reports whether the file allows no inline-tool findings.
func (r Row) Locked() bool { return r.Baseline == 0 }

// Regressed reports whether the current count exceeds the baseline.
func (r Row) Regressed() bool { return r.Current > r.Baseline }

// Improved reports whether the current count is below the baseline.
func (r Row) Improved() bool { return r.Current < r.Baseline }

// Counts returns the number of P001 findings that are not allowed in each
// scanned file, including files with none.
func Counts(paths []string, findings []Finding) map[string]int {
	counts := make(map[string]int, len(paths))
	for _, p := range paths {
		counts[p] = 0
	}
	for _, f := range findings {
		if f.Rule == InlineTool && f.Severity != SeverityAllowed {
			counts[f.File]++
		}
	}
	return counts
}

// Regressions returns the files whose new count is above the old baseline.
// A file missing from old counts as locked at 0.
func Regressions(old *Baseline, counts map[string]int) []Row {
	var rows []Row
	for p, n := range counts {
		if n > old.Workflows[p] {
			rows = append(rows, Row{File: p, Baseline: old.Workflows[p], Current: n})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].File < rows[j].File })
	return rows
}

// Apply marks the first baseline-count P001 findings of each file tolerated,
// adds one P004 finding per file over its baseline, and returns the sorted
// findings with one row per scanned file.
func Apply(b *Baseline, paths []string, findings []Finding) ([]Finding, []Row) {
	counts := Counts(paths, findings)
	seen := map[string]int{}
	for i := range findings {
		f := &findings[i]
		if f.Rule != InlineTool || f.Severity == SeverityAllowed {
			continue
		}
		if seen[f.File] < b.Workflows[f.File] {
			f.Severity = SeverityTolerated
		}
		seen[f.File]++
	}
	rows := make([]Row, 0, len(paths))
	for _, p := range paths {
		row := Row{File: p, Baseline: b.Workflows[p], Current: counts[p]}
		rows = append(rows, row)
		if row.Regressed() {
			findings = append(findings, regression(row))
		}
	}
	Sort(findings)
	sort.Slice(rows, func(i, j int) bool { return rows[i].File < rows[j].File })
	return findings, rows
}

func regression(r Row) Finding {
	msg := fmt.Sprintf("%s has %d inline-tool findings, baseline allows %d", r.File, r.Current, r.Baseline)
	if r.Locked() {
		msg = fmt.Sprintf("%s is locked at 0 and has %d inline-tool findings", r.File, r.Current)
	}
	return Finding{Rule: Regression, Severity: SeverityError, File: r.File, Line: 1, Col: 1, Message: msg}
}
