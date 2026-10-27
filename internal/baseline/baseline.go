// Package baseline reads, writes and applies the per-workflow count of
// tolerated inline-tool findings. Counts can only go down.
package baseline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jimmykurian/ci-parity/internal/rules"
)

// Version is the baseline file format version.
const Version = 1

// File is the content of .ci-parity-baseline.json.
type File struct {
	Version   int            `json:"version"`
	Workflows map[string]int `json:"workflows"`
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

// Read loads the baseline at path, relative to root. A missing file returns
// an error that wraps fs.ErrNotExist.
func Read(root, path string) (*File, error) {
	data, err := os.ReadFile(join(root, path))
	if err != nil {
		return nil, fmt.Errorf("read baseline %s: %w", filepath.ToSlash(path), err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var f File
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("baseline %s: %w", filepath.ToSlash(path), err)
	}
	if f.Version != Version {
		return nil, fmt.Errorf("baseline %s: unsupported version %d; want %d", filepath.ToSlash(path), f.Version, Version)
	}
	for wf, n := range f.Workflows {
		if n < 0 {
			return nil, fmt.Errorf("baseline %s: %s has negative count %d", filepath.ToSlash(path), wf, n)
		}
	}
	return &f, nil
}

// Write saves counts as a baseline at path, relative to root, with sorted keys
// and zero counts written out so locks show in diffs.
func Write(root, path string, counts map[string]int) error {
	data, err := json.MarshalIndent(File{Version: Version, Workflows: counts}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode baseline: %w", err)
	}
	if err := os.WriteFile(join(root, path), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write baseline %s: %w", filepath.ToSlash(path), err)
	}
	return nil
}

// Counts returns the number of P001 findings that are not allowed in each
// scanned file, including files with none.
func Counts(paths []string, findings []rules.Finding) map[string]int {
	counts := make(map[string]int, len(paths))
	for _, p := range paths {
		counts[p] = 0
	}
	for _, f := range findings {
		if f.Rule == rules.InlineTool && f.Severity != rules.SeverityAllowed {
			counts[f.File]++
		}
	}
	return counts
}

// Regressions returns the files whose new count is above the old baseline.
// A file missing from old counts as locked at 0.
func Regressions(old *File, counts map[string]int) []Row {
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
func Apply(b *File, paths []string, findings []rules.Finding) ([]rules.Finding, []Row) {
	counts := Counts(paths, findings)
	seen := map[string]int{}
	for i := range findings {
		f := &findings[i]
		if f.Rule != rules.InlineTool || f.Severity == rules.SeverityAllowed {
			continue
		}
		if seen[f.File] < b.Workflows[f.File] {
			f.Severity = rules.SeverityTolerated
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
	rules.Sort(findings)
	sort.Slice(rows, func(i, j int) bool { return rows[i].File < rows[j].File })
	return findings, rows
}

func regression(r Row) rules.Finding {
	msg := fmt.Sprintf("%s has %d inline-tool findings, baseline allows %d", r.File, r.Current, r.Baseline)
	if r.Locked() {
		msg = fmt.Sprintf("%s is locked at 0 and has %d inline-tool findings", r.File, r.Current)
	}
	return rules.Finding{Rule: rules.Regression, Severity: rules.SeverityError, File: r.File, Line: 1, Col: 1, Message: msg}
}

func join(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}
