// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/cli"
)

func TestFix(t *testing.T) {
	const fixture = "../../testdata/fix/basic"
	workflow := ".github/workflows/ci.yml"
	repo := copyDir(t, filepath.Join(fixture, "repo"))
	before := readFile(t, filepath.Join(repo, workflow))

	code, out, _ := runCLI("fix", "--dry-run", "-C", repo)
	// A dry run can't know the escaped value will fail the editor's check, so
	// it plans 9 edits; the real run makes 8 and counts that one as manual.
	wantDry := ".github/workflows/ci.yml:9:14 'pnpm install --frozen-lockfile' -> 'task install'\n" +
		".github/workflows/ci.yml:11:14 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:13:19 'pnpm lint' -> 'task lint'\n" +
		".github/workflows/ci.yml:15:43 'pnpm build' -> 'task build'\n" +
		".github/workflows/ci.yml:17:28 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:19:15 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:21:15 'pnpm lint' -> 'task lint'\n" +
		".github/workflows/ci.yml:24:11 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:28:15 'pnpm\ttest' -> 'task test'\n" +
		"would fix 9 steps in 1 file; 7 inline-tool findings need a manual change\n"
	if code != cli.ExitOK || out != wantDry {
		t.Errorf("dry run: exit %d, output:\n%s\nwant:\n%s", code, out, wantDry)
	}
	if readFile(t, filepath.Join(repo, workflow)) != before {
		t.Fatal("--dry-run changed the workflow")
	}

	code, out, _ = runCLI("fix", "-C", repo)
	want := ".github/workflows/ci.yml:9:14 'pnpm install --frozen-lockfile' -> 'task install'\n" +
		".github/workflows/ci.yml:11:14 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:13:19 'pnpm lint' -> 'task lint'\n" +
		".github/workflows/ci.yml:15:43 'pnpm build' -> 'task build'\n" +
		".github/workflows/ci.yml:17:28 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:19:15 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:21:15 'pnpm lint' -> 'task lint'\n" +
		".github/workflows/ci.yml:24:11 'pnpm test' -> 'task test'\n" +
		"fixed 8 steps in 1 file; 8 inline-tool findings need a manual change\n"
	if code != cli.ExitOK || out != want {
		t.Errorf("fix: exit %d, output:\n%s\nwant:\n%s", code, out, want)
	}
	if diff := cmp.Diff(readFile(t, filepath.Join(fixture, "expected", workflow)), readFile(t, filepath.Join(repo, workflow))); diff != "" {
		t.Errorf("fixed workflow (-want +got):\n%s", diff)
	}

	_, out, _ = runCLI("fix", "-C", repo)
	if want := "fixed 0 steps in 0 files; 8 inline-tool findings need a manual change\n"; out != want {
		t.Errorf("second run output:\n%s\nwant:\n%s", out, want)
	}
}

func TestFixSuggestsBaseline(t *testing.T) {
	repo := copyDir(t, "../../testdata/fix/basic/repo")
	if code, _, _ := runCLI("baseline", "-C", repo); code != cli.ExitOK {
		t.Fatalf("baseline exit %d", code)
	}
	_, out, _ := runCLI("fix", "-C", repo)
	if !bytes.Contains([]byte(out), []byte("run `ci-parity baseline` to lock in the lower counts")) {
		t.Errorf("no baseline hint in:\n%s", out)
	}
}
