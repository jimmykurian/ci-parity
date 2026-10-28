// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package testutil holds the golden fixture harness shared by CLI tests. Its
// comparisons take a T, which both *testing.T and Ginkgo's GinkgoT() satisfy.
package testutil

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files and exit codes from current output")

// Formats lists every output format the CLI supports.
var Formats = []string{"text", "github", "json", "sarif"}

// RunFunc runs the CLI in-process with args and returns its exit code.
type RunFunc func(args []string, stdout, stderr io.Writer) int

// T is the part of a test the comparisons use: *testing.T and Ginkgo's
// GinkgoT() both provide it.
type T interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// FixtureDirs returns the fixture directories directly under root, sorted by
// name. A missing root yields no fixtures. It needs no test, so Ginkgo tables
// can be built from it before any spec runs.
func FixtureDirs(root string) ([]string, error) {
	// Stat first: on Windows, reading a file as a directory looks like a missing one.
	info, err := os.Stat(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err == nil && !info.IsDir() {
		return nil, fmt.Errorf("read fixtures %s: not a directory", root)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read fixtures %s: %w", root, err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(root, e.Name()))
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

// Fixtures is FixtureDirs for a test, failing it when root can't be read.
func Fixtures(t T, root string) []string {
	t.Helper()
	dirs, err := FixtureDirs(root)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return dirs
}

// RunGolden runs `check -C <fixture> --format <format>` for every fixture under
// root in each of formats and compares stdout with expected/<format>.golden and
// the exit code with expected/exitcode. Every format must exit with the same
// code. With -update the golden files and exit code are rewritten.
func RunGolden(t *testing.T, root string, formats []string, run RunFunc) {
	t.Helper()
	dirs := Fixtures(t, root)
	if len(dirs) == 0 {
		t.Logf("no fixtures under %s", root)
		return
	}
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			for _, format := range formats {
				t.Run(format, func(t *testing.T) {
					var stdout, stderr bytes.Buffer
					code := run([]string{"check", "-C", dir, "--format", format}, &stdout, &stderr)
					expected := filepath.Join(dir, "expected")
					CompareGolden(t, filepath.Join(expected, format+".golden"), stdout.Bytes(), stderr.String())
					CompareExitCode(t, filepath.Join(expected, "exitcode"), code, stderr.String())
				})
			}
		})
	}
}

// CompareGolden fails the test when got differs from the golden file at path, or
// rewrites the file under -update. stderr is shown on a mismatch.
func CompareGolden(t T, path string, got []byte, stderr string) {
	t.Helper()
	if *update {
		writeFile(t, path, got)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run with -update to create): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s mismatch (run with -update to rewrite)\n--- want\n%s\n--- got\n%s\n--- stderr\n%s", path, want, got, stderr)
	}
}

// CompareExitCode fails the test when got differs from the exit code in path, or
// rewrites it under -update. stderr is shown on a mismatch.
func CompareExitCode(t T, path string, got int, stderr string) {
	t.Helper()
	if *update {
		writeFile(t, path, []byte(strconv.Itoa(got)+"\n"))
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exit code (run with -update to create): %v", err)
	}
	want, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if got != want {
		t.Errorf("exit code = %d, want %d\n--- stderr\n%s", got, want, stderr)
	}
}

// writeFile writes data to path, creating its directory, and fails the test on
// error.
func writeFile(t T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
