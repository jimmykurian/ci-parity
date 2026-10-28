// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package jsonfile reads and writes .ci-parity-baseline.json.
package jsonfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Version is the baseline file format version.
const Version = 1

// file is the content of .ci-parity-baseline.json.
type file struct {
	Version   int            `json:"version"`
	Workflows map[string]int `json:"workflows"`
}

// Read loads the baseline at path, relative to root. A missing file returns
// an error that wraps fs.ErrNotExist.
func Read(root, path string) (*domain.Baseline, error) {
	data, err := os.ReadFile(join(root, path))
	if err != nil {
		return nil, fmt.Errorf("read baseline %s: %w", filepath.ToSlash(path), err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var f file
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
	return &domain.Baseline{Workflows: f.Workflows}, nil
}

// Write saves counts as a baseline at path, relative to root, with sorted keys
// and zero counts written out so locks show in diffs.
func Write(root, path string, counts map[string]int) error {
	data, err := json.MarshalIndent(file{Version: Version, Workflows: counts}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode baseline: %w", err)
	}
	//nolint:gosec // G306: the baseline is committed and read by everyone, like source.
	if err := os.WriteFile(join(root, path), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write baseline %s: %w", filepath.ToSlash(path), err)
	}
	return nil
}

// join resolves path against root unless it is already absolute.
func join(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}
