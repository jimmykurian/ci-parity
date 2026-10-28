// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// fixCopy copies the fixture repo at src into a new temporary directory.
func fixCopy(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

func fixCLI(t *testing.T, args ...string) (string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	if stderr.Len() > 0 {
		t.Logf("stderr: %s", stderr.String())
	}
	return stdout.String(), code
}

func fixRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestFix(t *testing.T) {
	const fixture = "../../testdata/fix/basic"
	workflow := ".github/workflows/ci.yml"
	repo := fixCopy(t, filepath.Join(fixture, "repo"))
	before := fixRead(t, filepath.Join(repo, workflow))

	out, code := fixCLI(t, "fix", "--dry-run", "-C", repo)
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
	if code != exitOK || out != wantDry {
		t.Errorf("dry run: exit %d, output:\n%s\nwant:\n%s", code, out, wantDry)
	}
	if fixRead(t, filepath.Join(repo, workflow)) != before {
		t.Fatal("--dry-run changed the workflow")
	}

	out, code = fixCLI(t, "fix", "-C", repo)
	want := ".github/workflows/ci.yml:9:14 'pnpm install --frozen-lockfile' -> 'task install'\n" +
		".github/workflows/ci.yml:11:14 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:13:19 'pnpm lint' -> 'task lint'\n" +
		".github/workflows/ci.yml:15:43 'pnpm build' -> 'task build'\n" +
		".github/workflows/ci.yml:17:28 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:19:15 'pnpm test' -> 'task test'\n" +
		".github/workflows/ci.yml:21:15 'pnpm lint' -> 'task lint'\n" +
		".github/workflows/ci.yml:24:11 'pnpm test' -> 'task test'\n" +
		"fixed 8 steps in 1 file; 8 inline-tool findings need a manual change\n"
	if code != exitOK || out != want {
		t.Errorf("fix: exit %d, output:\n%s\nwant:\n%s", code, out, want)
	}
	if diff := cmp.Diff(fixRead(t, filepath.Join(fixture, "expected", workflow)), fixRead(t, filepath.Join(repo, workflow))); diff != "" {
		t.Errorf("fixed workflow (-want +got):\n%s", diff)
	}

	out, _ = fixCLI(t, "fix", "-C", repo)
	if want := "fixed 0 steps in 0 files; 8 inline-tool findings need a manual change\n"; out != want {
		t.Errorf("second run output:\n%s\nwant:\n%s", out, want)
	}
}

func TestFixSuggestsBaseline(t *testing.T) {
	repo := fixCopy(t, "../../testdata/fix/basic/repo")
	if _, code := fixCLI(t, "baseline", "-C", repo); code != exitOK {
		t.Fatalf("baseline exit %d", code)
	}
	out, _ := fixCLI(t, "fix", "-C", repo)
	if !bytes.Contains([]byte(out), []byte("run `ci-parity baseline` to lock in the lower counts")) {
		t.Errorf("no baseline hint in:\n%s", out)
	}
}
