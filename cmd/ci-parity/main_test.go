package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jimmykurian/ci-parity/internal/testutil"
)

func TestGolden(t *testing.T) {
	testutil.RunGolden(t, "../../testdata/fixtures", testutil.Formats, run)
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
		{"check unknown format", []string{"check", "--format", "sarif"}, exitUsage, "", `unknown format "sarif"`},
		{"check missing explicit config", []string{"check", "-C", "testdata-absent", "--config", "x.yml"}, exitUsage, "", "read config"},
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
