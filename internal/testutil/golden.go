// Package testutil holds the golden fixture harness shared by CLI tests.
package testutil

import (
	"bytes"
	"errors"
	"flag"
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

// Formats lists the output formats every fixture is checked in.
var Formats = []string{"text", "github", "json"}

// RunFunc runs the CLI in-process with args and returns its exit code.
type RunFunc func(args []string, stdout, stderr io.Writer) int

// Fixtures returns the fixture directories directly under root, sorted by name.
// A missing root yields no fixtures.
func Fixtures(t testing.TB, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read fixtures %s: %v", root, err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(root, e.Name()))
		}
	}
	sort.Strings(dirs)
	return dirs
}

// RunGolden runs `check -C <fixture> --format <format>` for every fixture under
// root in every format and compares stdout with expected/<format>.golden and the
// exit code with expected/exitcode. A golden file that does not exist yet is
// skipped; with -update every golden file and exit code is rewritten.
func RunGolden(t *testing.T, root string, run RunFunc) {
	t.Helper()
	dirs := Fixtures(t, root)
	if len(dirs) == 0 {
		t.Logf("no fixtures under %s", root)
		return
	}
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			for _, format := range Formats {
				t.Run(format, func(t *testing.T) {
					var stdout, stderr bytes.Buffer
					code := run([]string{"check", "-C", dir, "--format", format}, &stdout, &stderr)
					expected := filepath.Join(dir, "expected")
					compareGolden(t, filepath.Join(expected, format+".golden"), stdout.Bytes(), stderr.String())
					compareExitCode(t, filepath.Join(expected, "exitcode"), code, stderr.String())
				})
			}
		})
	}
}

func compareGolden(t *testing.T, path string, got []byte, stderr string) {
	t.Helper()
	if *update {
		writeFile(t, path, got)
		return
	}
	want, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		t.Skipf("no golden file %s", path)
	}
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s mismatch (run with -update to rewrite)\n--- want\n%s\n--- got\n%s\n--- stderr\n%s", path, want, got, stderr)
	}
}

func compareExitCode(t *testing.T, path string, got int, stderr string) {
	t.Helper()
	if *update {
		writeFile(t, path, []byte(strconv.Itoa(got)+"\n"))
		return
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	if err != nil {
		t.Fatalf("read exit code: %v", err)
	}
	want, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if got != want {
		t.Errorf("exit code = %d, want %d\n--- stderr\n%s", got, want, stderr)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
