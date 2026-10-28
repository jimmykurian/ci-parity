// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package e2e builds the ci-parity binary once and runs it as a separate
// process, checking what only a real process shows: the version stamped in with
// -ldflags, exit codes, the split between stdout and stderr, the working
// directory standing in for -C, and files written to disk.
package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fixtures holds the golden fixture repositories, relative to this package.
const fixtures = "../../testdata/fixtures"

// stamp is the build metadata the binary is built with, as `version` prints it.
const stamp = "ci-parity e2e (abc1234, 2026-01-02T03:04:05Z)"

// binary is the path of the ci-parity binary TestMain builds.
var binary string

// TestMain builds the binary into a temporary directory, runs the tests, and
// removes it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "ci-parity-e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e:", err)
		os.Exit(1)
	}
	code := run(m, dir)
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// run builds the binary into dir and runs the tests, returning their exit code.
func run(m *testing.M, dir string) int {
	binary = filepath.Join(dir, "ci-parity")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	pkg := "github.com/jimmykurian/ci-parity/internal/cli"
	ldflags := fmt.Sprintf("-X %[1]s.version=e2e -X %[1]s.commit=abc1234 -X %[1]s.date=2026-01-02T03:04:05Z", pkg)
	build := exec.Command("go", "build", "-ldflags", ldflags, "-o", binary, "../../cmd/ci-parity")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: build ci-parity: %v\n%s", err, out)
		return 1
	}
	return m.Run()
}

// result is what one run of the binary produced.
type result struct {
	code           int
	stdout, stderr string
}

// ciParity runs the binary with args in dir and returns its exit code and
// output. It fails the test when the process can't be started.
func ciParity(t *testing.T, dir string, args ...string) result {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatalf("run ci-parity %v: %v", args, err)
	}
	return result{code: cmd.ProcessState.ExitCode(), stdout: stdout.String(), stderr: stderr.String()}
}

// golden returns a fixture's expected output for format.
func golden(t *testing.T, fixture, format string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixtures, fixture, "expected", format+".golden"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestVersionCarriesTheBuildStamp(t *testing.T) {
	r := ciParity(t, ".", "version")
	if r.code != 0 || r.stdout != stamp+"\n" || r.stderr != "" {
		t.Errorf("version: exit %d, stdout %q, stderr %q; want exit 0 and %q", r.code, r.stdout, r.stderr, stamp)
	}
}

func TestExitCodesAndStreams(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string // a substring; "" means stdout must be empty
		wantStderr string // a substring; "" means stderr must be empty
	}{
		{"clean repository", []string{"check", "-C", filepath.Join(fixtures, "clean")}, 0, "0 errors", ""},
		{"error findings", []string{"check", "-C", filepath.Join(fixtures, "inline-tools")}, 1, "P001 inline-tool", ""},
		{"help", []string{"--help"}, 0, "Usage:", ""},
		{"no command", nil, 2, "", "Usage:"},
		{"unknown command", []string{"lint"}, 2, "", `unknown command "lint"`},
		{"missing directory", []string{"check", "-C", "absent"}, 2, "", "-C absent: not a directory"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ciParity(t, ".", tt.args...)
			if r.code != tt.wantCode {
				t.Errorf("exit %d, want %d (stderr %q)", r.code, tt.wantCode, r.stderr)
			}
			for _, s := range []struct{ name, got, want string }{{"stdout", r.stdout, tt.wantStdout}, {"stderr", r.stderr, tt.wantStderr}} {
				if s.want == "" && s.got != "" {
					t.Errorf("%s = %q, want it empty", s.name, s.got)
				}
				if !strings.Contains(s.got, s.want) {
					t.Errorf("%s = %q, want it to contain %q", s.name, s.got, s.want)
				}
			}
		})
	}
}

func TestWorkingDirectoryStandsInForC(t *testing.T) {
	for _, format := range []string{"text", "json"} {
		t.Run(format, func(t *testing.T) {
			r := ciParity(t, filepath.Join(fixtures, "inline-tools"), "check", "--format", format)
			if want := golden(t, "inline-tools", format); r.code != 1 || r.stdout != want {
				t.Errorf("exit %d, stdout:\n%s\nwant exit 1 and:\n%s", r.code, r.stdout, want)
			}
		})
	}
}

func TestFixRewritesTheWorkflow(t *testing.T) {
	src := "../../testdata/fix/basic"
	repo := t.TempDir()
	if err := os.CopyFS(repo, os.DirFS(filepath.Join(src, "repo"))); err != nil {
		t.Fatal(err)
	}
	if r := ciParity(t, repo, "fix"); r.code != 0 || !strings.Contains(r.stdout, "fixed 8 steps in 1 file") {
		t.Fatalf("fix: exit %d, stdout %q, stderr %q", r.code, r.stdout, r.stderr)
	}
	workflow := filepath.Join(".github", "workflows", "ci.yml")
	got, err := os.ReadFile(filepath.Join(repo, workflow))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(src, "expected", workflow))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("fixed workflow:\n%s\nwant:\n%s", got, want)
	}
}
