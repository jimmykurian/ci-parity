// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"version", []string{"version"}, ExitOK, "ci-parity dev (none, unknown)\n", ""},
		{"no args", nil, ExitUsage, "", "Usage:"},
		{"help", []string{"--help"}, ExitOK, "Usage:", ""},
		{"unknown command", []string{"lint"}, ExitUsage, "", `unknown command "lint"`},
		{"check unknown format", []string{"check", "--format", "xml"}, ExitUsage, "", `unknown format "xml"`},
		{"check missing directory", []string{"check", "-C", "testdata-absent"}, ExitUsage, "", "-C testdata-absent: not a directory"},
		{"check missing explicit config", []string{"check", "-C", ".", "--config", "x.yml"}, ExitUsage, "", "read config"},
		{"unknown flag", []string{"check", "--nope"}, ExitUsage, "", "flag provided but not defined"},
		{"positional argument", []string{"version", "extra"}, ExitUsage, "", `unexpected argument "extra"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
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
