// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package ghactions

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Apply makes the edits to the repo-relative file at path and returns those it
// made. Columns count characters. An edit whose Old text is not at its
// position, because the file changed since it was read, is skipped. Line
// endings and the file's mode are kept.
func (s Source) Apply(path string, edits []domain.Edit) ([]domain.Edit, error) {
	full := filepath.Join(s.Root, filepath.FromSlash(path))
	if rel, err := filepath.Rel(s.Root, full); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("fix %s: outside the repository", path)
	}
	info, err := os.Stat(full)
	if err != nil {
		return nil, fmt.Errorf("fix %s: %w", path, err)
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("fix %s: %w", path, err)
	}
	lines := strings.SplitAfter(string(data), "\n")
	// Right to left within a line, so earlier columns stay valid.
	sorted := append([]domain.Edit(nil), edits...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Line != sorted[j].Line {
			return sorted[i].Line < sorted[j].Line
		}
		return sorted[i].Col > sorted[j].Col
	})
	var applied []domain.Edit
	for _, e := range sorted {
		if e.Line < 1 || e.Line > len(lines) {
			continue
		}
		line := []rune(lines[e.Line-1])
		start, old := e.Col-1, []rune(e.Old)
		if start < 0 || start+len(old) > len(line) || string(line[start:start+len(old)]) != e.Old {
			continue
		}
		lines[e.Line-1] = string(line[:start]) + e.New + string(line[start+len(old):])
		applied = append(applied, e)
	}
	if len(applied) == 0 {
		return nil, nil
	}
	//nolint:gosec // G703: full is a discovered workflow path, checked above to be inside Root.
	if err := os.WriteFile(full, []byte(strings.Join(lines, "")), info.Mode().Perm()); err != nil {
		return nil, fmt.Errorf("fix %s: %w", path, err)
	}
	sort.Slice(applied, func(i, j int) bool {
		if applied[i].Line != applied[j].Line {
			return applied[i].Line < applied[j].Line
		}
		return applied[i].Col < applied[j].Col
	})
	return applied, nil
}
