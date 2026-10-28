// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package ghactions

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

func TestApply(t *testing.T) {
	// Arrange
	root := t.TempDir()
	path := filepath.Join(root, "ci.yml")
	in := "steps:\r\n  - run: pnpm test # note\r\n  - run: pnpm lint\r\n  - run: énv pnpm x\r\n"
	if err := os.WriteFile(path, []byte(in), 0o600); err != nil {
		t.Fatal(err)
	}
	edits := []domain.Edit{
		{File: "ci.yml", Line: 2, Col: 10, Old: "pnpm test", New: "task test"},
		{File: "ci.yml", Line: 3, Col: 10, Old: "pnpm build", New: "task build"}, // text moved: skipped
		{File: "ci.yml", Line: 4, Col: 14, Old: "pnpm x", New: "task x"},         // after a multibyte character; yaml.v3 columns count characters
		{File: "ci.yml", Line: 9, Col: 1, Old: "x", New: "y"},                    // no such line
	}
	wantApplied := []domain.Edit{edits[0], edits[2]}
	want := "steps:\r\n  - run: task test # note\r\n  - run: pnpm lint\r\n  - run: énv task x\r\n"

	// Act
	applied, err := Source{Root: root}.Apply("ci.yml", edits)
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if diff := cmp.Diff(wantApplied, applied); diff != "" {
		t.Errorf("applied (-want +got):\n%s", diff)
	}
	got, _ := os.ReadFile(path)
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("file (-want +got):\n%s", diff)
	}
	// Windows has no Unix permission bits to keep, so the mode is checked elsewhere.
	if info, _ := os.Stat(path); runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestApplyRefusesPathsOutsideRoot(t *testing.T) {
	// Arrange
	root := t.TempDir()
	src := Source{Root: filepath.Join(root, "repo")}
	edits := []domain.Edit{{Line: 1, Col: 1, Old: "a", New: "b"}}

	// Act
	_, err := src.Apply("../outside.yml", edits)

	// Assert
	if err == nil {
		t.Error("an edit outside the repository was accepted")
	}
}
