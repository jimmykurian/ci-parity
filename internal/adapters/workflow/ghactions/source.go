// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package ghactions

import "github.com/jimmykurian/ci-parity/internal/domain"

// Source implements app.WorkflowSource for the repository at Root.
type Source struct {
	Root string
}

// Discover returns the repo-relative paths matching any of patterns, sorted.
func (s Source) Discover(patterns []string) ([]string, error) {
	return Discover(s.Root, patterns)
}

// Load parses the workflow or action at the repo-relative path.
func (s Source) Load(path string) (*domain.File, error) {
	return Load(s.Root, path)
}
