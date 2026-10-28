// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package jsonfile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadMissingFile(t *testing.T) {
	// Arrange
	root := t.TempDir()

	// Act
	_, err := Read(root, "missing.json")

	// Assert
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing file error = %v, want fs.ErrNotExist", err)
	}
}

func TestReadErrors(t *testing.T) {
	// Arrange
	root := t.TempDir()
	tests := map[string]string{
		`{"version": 2, "workflows": {}}`:             "unsupported version 2",
		`{"version": 1, "workflows": {"a.yml": -1}}`:  "negative count",
		`{"version": 1, "workflows": {}, "extra": 1}`: "unknown field",
		`{"version": 1, "workflows": {"a.yml": "x"}}`: "cannot unmarshal",
	}
	for content, want := range tests {
		// Arrange
		if err := os.WriteFile(filepath.Join(root, "b.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		// Act
		_, err := Read(root, "b.json")

		// Assert
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Read(%s) error = %v, want %q", content, err, want)
		}
	}
}
