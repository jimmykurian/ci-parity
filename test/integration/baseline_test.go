// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jimmykurian/ci-parity/internal/cli"
)

func TestBaselineRefusesRegression(t *testing.T) {
	dir := copyDir(t, filepath.Join(fixtures, "baseline-regression"))
	path := filepath.Join(dir, ".ci-parity-baseline.json")
	before := readFile(t, path)

	code, _, stderr := runCLI("baseline", "-C", dir)
	if code != cli.ExitFindings || !strings.Contains(stderr, ".github/workflows/ci.yml  1 -> 3") {
		t.Fatalf("baseline = %d, stderr %q; want refusal listing the workflow", code, stderr)
	}
	if readFile(t, path) != before {
		t.Error("refused baseline must not change the file")
	}

	code, stdout, stderr := runCLI("baseline", "-C", dir, "--allow-regression")
	if code != cli.ExitOK {
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
	dir := copyDir(t, filepath.Join(fixtures, "baseline-locked"))
	if err := os.Remove(filepath.Join(dir, ".ci-parity-baseline.json")); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runCLI("baseline", "-C", dir); code != cli.ExitOK {
		t.Fatalf("first baseline = %d, stderr %q", code, stderr)
	}
	want := "{\n  \"version\": 1,\n  \"workflows\": {\n" +
		"    \".github/workflows/ci.yml\": 3,\n" +
		"    \".github/workflows/nightly.yml\": 1,\n" +
		"    \".github/workflows/release.yml\": 1\n  }\n}\n"
	if got := readFile(t, filepath.Join(dir, ".ci-parity-baseline.json")); got != want {
		t.Errorf("baseline file =\n%s\nwant\n%s", got, want)
	}
	if code, stdout, _ := runCLI("check", "-C", dir); code != cli.ExitOK || !strings.Contains(stdout, "5 tolerated") {
		t.Errorf("check after baseline = %d, %q; want everything tolerated", code, stdout)
	}
}

func TestCheckBaselineFlags(t *testing.T) {
	dir := filepath.Join(fixtures, "baseline-within")
	tests := []struct {
		name     string
		args     []string
		wantCode int
		want     string
	}{
		{"baseline applies", nil, cli.ExitOK, "3 tolerated"},
		{"strict ignores it", []string{"--strict"}, cli.ExitFindings, "3 errors, 0 warnings, 0 tolerated"},
		{"verbose shows tolerated", []string{"--verbose"}, cli.ExitOK, "runs 'pnpm install' directly [tolerated]"},
		{"explicit missing path is strict", []string{"--baseline", "none.json"}, cli.ExitFindings, "3 errors"},
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
