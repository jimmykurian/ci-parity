package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copyFixture copies a fixture repo into a temp dir so tests can write to it.
func copyFixture(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join("../../testdata/fixtures", name)
	dst := t.TempDir()
	err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

func runCLI(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestBaselineRefusesRegression(t *testing.T) {
	dir := copyFixture(t, "baseline-regression")
	path := filepath.Join(dir, ".ci-parity-baseline.json")
	before := readFile(t, path)

	code, _, stderr := runCLI("baseline", "-C", dir)
	if code != exitFindings || !strings.Contains(stderr, ".github/workflows/ci.yml  1 -> 3") {
		t.Fatalf("baseline = %d, stderr %q; want refusal listing the workflow", code, stderr)
	}
	if readFile(t, path) != before {
		t.Error("refused baseline must not change the file")
	}

	code, stdout, stderr := runCLI("baseline", "-C", dir, "--allow-regression")
	if code != exitOK {
		t.Fatalf("baseline --allow-regression = %d, stderr %q", code, stderr)
	}
	want := "{\n  \"version\": 1,\n  \"workflows\": {\n    \".github/workflows/ci.yml\": 3\n  }\n}\n"
	if got := readFile(t, path); got != want {
		t.Errorf("baseline file =\n%s\nwant\n%s", got, want)
	}
	if !strings.Contains(stdout, "1 workflow, 3 inline-tool findings") {
		t.Errorf("stdout = %q", stdout)
	}
}

func TestBaselineCreatesThenLocksIn(t *testing.T) {
	dir := copyFixture(t, "baseline-locked")
	if err := os.Remove(filepath.Join(dir, ".ci-parity-baseline.json")); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runCLI("baseline", "-C", dir); code != exitOK {
		t.Fatalf("first baseline = %d, stderr %q", code, stderr)
	}
	want := "{\n  \"version\": 1,\n  \"workflows\": {\n" +
		"    \".github/workflows/ci.yml\": 3,\n" +
		"    \".github/workflows/nightly.yml\": 1,\n" +
		"    \".github/workflows/release.yml\": 1\n  }\n}\n"
	if got := readFile(t, filepath.Join(dir, ".ci-parity-baseline.json")); got != want {
		t.Errorf("baseline file =\n%s\nwant\n%s", got, want)
	}
	if code, stdout, _ := runCLI("check", "-C", dir); code != exitOK || !strings.Contains(stdout, "5 tolerated") {
		t.Errorf("check after baseline = %d, %q; want everything tolerated", code, stdout)
	}
}

func TestCheckBaselineFlags(t *testing.T) {
	dir := "../../testdata/fixtures/baseline-within"
	tests := []struct {
		name     string
		args     []string
		wantCode int
		want     string
	}{
		{"baseline applies", nil, exitOK, "3 tolerated"},
		{"strict ignores it", []string{"--strict"}, exitFindings, "3 errors, 0 warnings, 0 tolerated"},
		{"verbose shows tolerated", []string{"--verbose"}, exitOK, "runs 'pnpm install' directly [tolerated]"},
		{"explicit missing path is strict", []string{"--baseline", "none.json"}, exitFindings, "3 errors"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(append([]string{"check", "-C", dir}, tt.args...)...)
			if code != tt.wantCode || !strings.Contains(stdout, tt.want) {
				t.Errorf("exit %d, stdout %q, stderr %q; want exit %d containing %q", code, stdout, stderr, tt.wantCode, tt.want)
			}
		})
	}
}
