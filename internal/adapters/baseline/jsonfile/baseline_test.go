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

func TestReadErrors(t *testing.T) {
	root := t.TempDir()
	if _, err := Read(root, "missing.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing file error = %v, want fs.ErrNotExist", err)
	}
	tests := map[string]string{
		`{"version": 2, "workflows": {}}`:             "unsupported version 2",
		`{"version": 1, "workflows": {"a.yml": -1}}`:  "negative count",
		`{"version": 1, "workflows": {}, "extra": 1}`: "unknown field",
		`{"version": 1, "workflows": {"a.yml": "x"}}`: "cannot unmarshal",
	}
	for content, want := range tests {
		if err := os.WriteFile(filepath.Join(root, "b.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Read(root, "b.json"); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Read(%s) error = %v, want %q", content, err, want)
		}
	}
}
