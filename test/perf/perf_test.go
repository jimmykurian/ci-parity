// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package perf measures ci-parity check on a generated repository with 50 workflows, the
// size the docs promise checks in well under a second.
package perf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jimmykurian/ci-parity/internal/cli"
)

const (
	workflows    = 50
	jobs         = 3
	stepsPerJob  = 8
	tasks        = 20
	maxCheckTime = time.Second
)

// writeRepo writes a repository with a Taskfile of tasks and workflows of jobs x
// stepsPerJob steps: task calls, inline tools with and without a matching task,
// wrappers, and multi-line scripts.
func writeRepo(tb testing.TB) string {
	tb.Helper()
	root := tb.TempDir()
	var taskfile strings.Builder
	taskfile.WriteString("version: \"3\"\n\ntasks:\n")
	for i := range tasks {
		fmt.Fprintf(&taskfile, "  task-%d:\n    cmds:\n      - pnpm run script-%d\n", i, i)
	}
	write(tb, filepath.Join(root, "Taskfile.yml"), taskfile.String())

	steps := []string{
		"run: task task-%d",
		"run: pnpm run script-%d",
		"run: sudo -E env CI=1 pnpm run script-%d",
		"run: |\n          echo step %d\n          go test ./...",
		"run: npx eslint . # %d",
		"run: task task-%d -- --verbose",
		"run: docker build -t acme/widget-service:%d .",
		"run: echo done %d",
	}
	for w := range workflows {
		var wf strings.Builder
		wf.WriteString("on: [push]\njobs:\n")
		for j := range jobs {
			fmt.Fprintf(&wf, "  job-%d:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v7\n", j)
			for s := range stepsPerJob {
				fmt.Fprintf(&wf, "      - name: step %d\n        "+steps[s]+"\n", s, (w+j+s)%tasks)
			}
		}
		write(tb, filepath.Join(root, ".github", "workflows", fmt.Sprintf("ci-%02d.yml", w)), wf.String())
	}
	return root
}

// write creates path with its directories and writes data to it.
func write(tb testing.TB, path, data string) {
	tb.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		tb.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		tb.Fatal(err)
	}
}

// check runs ci-parity check on root and returns its exit code and stdout.
func check(root string) (int, string) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"check", "-C", root, "--format", "json"}, &stdout, &stderr)
	return code, stdout.String()
}

func TestCheck50WorkflowsUnderASecond(t *testing.T) {
	// Arrange
	root := writeRepo(t)
	check(root) // warm up the file cache and the runtime

	// Act
	start := time.Now()
	code, out := check(root)
	took := time.Since(start)

	// Assert
	if code != cli.ExitFindings || !strings.Contains(out, `"rule": "P001"`) {
		t.Fatalf("check exit %d; want %d with P001 findings, so the time measures real work\n%.500s", code, cli.ExitFindings, out)
	}
	if took > maxCheckTime {
		t.Errorf("check on %d workflows took %v, want under %v", workflows, took, maxCheckTime)
	}
	t.Logf("check on %d workflows (%d steps) took %v", workflows, workflows*jobs*stepsPerJob, took)
}

func BenchmarkCheck50Workflows(b *testing.B) {
	root := writeRepo(b)
	b.ReportAllocs()
	for b.Loop() {
		check(root)
	}
}
