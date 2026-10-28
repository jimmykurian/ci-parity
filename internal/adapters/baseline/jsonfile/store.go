// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package jsonfile

import "github.com/jimmykurian/ci-parity/internal/domain"

// Store implements app.BaselineStore for the repository at Root.
type Store struct {
	Root string
}

// Read loads the baseline at path; a missing file wraps fs.ErrNotExist.
func (s Store) Read(path string) (*domain.Baseline, error) {
	return Read(s.Root, path)
}

// Write saves counts as the baseline at path.
func (s Store) Write(path string, counts map[string]int) error {
	return Write(s.Root, path, counts)
}
