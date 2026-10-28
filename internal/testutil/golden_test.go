// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureDirsReportsUnreadableRoot(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := FixtureDirs(file); err == nil {
		t.Error("FixtureDirs on a file: want an error")
	}
}

func TestFixturesMissingRoot(t *testing.T) {
	if got := Fixtures(t, filepath.Join(t.TempDir(), "absent")); got != nil {
		t.Errorf("Fixtures = %v, want nil", got)
	}
}

func TestFixturesSortedDirsOnly(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"widget-service", "acme-app"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".gitkeep"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got := Fixtures(t, root)
	want := []string{filepath.Join(root, "acme-app"), filepath.Join(root, "widget-service")}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("Fixtures = %v, want %v", got, want)
	}
}

func TestRunGoldenPassesMatchingFixture(t *testing.T) {
	root := t.TempDir()
	expected := filepath.Join(root, "example-api", "expected")
	if err := os.MkdirAll(expected, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, format := range Formats {
		if err := os.WriteFile(filepath.Join(expected, format+".golden"), []byte(format+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(expected, "exitcode"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var calls int
	RunGolden(t, root, Formats, func(args []string, stdout, _ io.Writer) int {
		calls++
		fmt.Fprintln(stdout, args[len(args)-1])
		return 1
	})
	if calls != len(Formats) {
		t.Errorf("run called %d times, want %d", calls, len(Formats))
	}
}
