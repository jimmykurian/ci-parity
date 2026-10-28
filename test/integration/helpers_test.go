// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package integration runs the ci-parity command line in-process, with the real
// adapters, against the fixture repositories in testdata/.
package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jimmykurian/ci-parity/internal/cli"
)

// fixtures holds the golden fixture repositories, relative to this package.
const fixtures = "../../testdata/fixtures"

// copyDir copies the fixture repo at src into a new temporary directory, so a
// test can write to it.
func copyDir(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

// runCLI runs ci-parity with args and returns its exit code, stdout and stderr.
func runCLI(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := cli.Run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// readFile returns the contents of path, failing the test when it can't be read.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
