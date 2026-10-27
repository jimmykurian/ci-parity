package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jimmykurian/ci-parity/internal/testutil"
)

// goldenFormats lists the formats fixtures are checked in; github and json join in Phase 3.
var goldenFormats = []string{"text"}

func TestGolden(t *testing.T) {
	testutil.RunGolden(t, "../../testdata/fixtures", goldenFormats, run)
}

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"version", []string{"version"}, exitOK, "ci-parity dev (none, unknown)\n", ""},
		{"no args", nil, exitUsage, "", "Usage:"},
		{"help", []string{"--help"}, exitOK, "Usage:", ""},
		{"unknown command", []string{"lint"}, exitUsage, "", `unknown command "lint"`},
		{"check not implemented", []string{"check", "--format", "json"}, exitUsage, "", "check: not implemented"},
		{"baseline not implemented", []string{"baseline", "--allow-regression"}, exitUsage, "", "baseline: not implemented"},
		{"unknown flag", []string{"check", "--nope"}, exitUsage, "", "flag provided but not defined"},
		{"positional argument", []string{"version", "extra"}, exitUsage, "", `unexpected argument "extra"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Errorf("stdout = %q, want it to contain %q", stdout.String(), tt.wantStdout)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}
