// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package mise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
	"github.com/jimmykurian/ci-parity/internal/rules"
)

var _ rules.TaskSource = (*Mise)(nil)

// write creates files; a name ending in "*" is written executable, without it.
func write(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		mode := os.FileMode(0o644)
		if strings.HasSuffix(name, "*") {
			name, mode = strings.TrimSuffix(name, "*"), 0o755
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const config = `[tools]
go = "1.27"

[tasks]
fmt = "gofmt -l ."
check = ["go vet ./...", "go test ./..."]

[tasks.build]
run = "go build ./..."
alias = "b"

[tasks."lint:go"]
run = ["golangci-lint run"]
alias = ["lg", "lint-go"]

[tasks.release]
run = "goreleaser release --clean"
hide = true

[task_config]
includes = ["tasks.toml", "more-tasks", "missing.toml"]
`

func load(t *testing.T) *Mise {
	t.Helper()
	root := write(t, map[string]string{
		".config/mise.toml":      "[tasks.build]\nrun = \"overridden\"\n[tasks.legacy]\nrun = \"old\"\n",
		"mise.toml":              config,
		"tasks.toml":             "[docs]\nrun = \"mkdocs build\"\n",
		"more-tasks/seed*":       "#!/bin/sh\npsql -f seed.sql\n",
		"mise-tasks/db/migrate*": "#!/bin/sh\ngoose up\n",
		"mise-tasks/secret*":     "#!/bin/sh\n#MISE hide=true\necho hidden\n",
		"mise-tasks/README.md":   "not executable, not a task\n",
	})
	m, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestTasks(t *testing.T) {
	// Arrange
	want := []string{"b", "build", "check", "db:migrate", "docs", "fmt", "legacy", "lg", "lint-go", "lint:go", "release", "secret", "seed"}
	commands := map[string][]string{
		"build":      {"go build ./..."}, // mise.toml wins over .config/mise.toml
		"check":      {"go vet ./...", "go test ./..."},
		"fmt":        {"gofmt -l ."},
		"lint:go":    {"golangci-lint run"},
		"docs":       {"mkdocs build"},
		"db:migrate": nil,
		"b":          nil,
	}
	wantHidden := map[string]bool{"release": true, "secret": true, "build": false, "db:migrate": false}

	// Act
	m := load(t)

	// Assert
	if diff := cmp.Diff(want, m.Tasks()); diff != "" {
		t.Errorf("tasks (-want +got):\n%s", diff)
	}
	for name, want := range commands {
		if diff := cmp.Diff(want, m.Commands(name)); diff != "" {
			t.Errorf("Commands(%s) (-want +got):\n%s", name, diff)
		}
	}
	for name, hidden := range wantHidden {
		if got := m.IsInternal(name); got != hidden {
			t.Errorf("IsInternal(%s) = %v, want %v", name, got, hidden)
		}
	}
	if m.File() != "mise.toml" || m.Command() != "mise run" || m.Internal() != "is hidden" {
		t.Errorf("File=%q Command=%q Internal=%q", m.File(), m.Command(), m.Internal())
	}
}

func TestLoadReportsMissingConfig(t *testing.T) {
	// Arrange
	root := write(t, map[string]string{"README.md": "x"})

	// Act
	_, err := Load(root)

	// Assert
	if err == nil || !strings.Contains(err.Error(), "no mise config found") {
		t.Errorf("no config: err = %v", err)
	}
}

func TestLoadReportsBadTOML(t *testing.T) {
	// Arrange
	root := write(t, map[string]string{"mise.toml": "[tasks\n"})

	// Act
	_, err := Load(root)

	// Assert
	if err == nil || !strings.Contains(err.Error(), "parse mise.toml") {
		t.Errorf("bad TOML: err = %v", err)
	}
}

func TestLoadReadsFileTasksWithoutConfig(t *testing.T) {
	// Arrange
	root := write(t, map[string]string{"mise-tasks/only*": "#!/bin/sh\n"})

	// Act
	m, err := Load(root)

	// Assert
	if err != nil || m.File() != "mise.toml" || len(m.Tasks()) != 1 {
		t.Errorf("file tasks only: m = %+v, err = %v", m, err)
	}
}

func TestNames(t *testing.T) {
	// Arrange
	words := func(s string) []domain.Word {
		var out []domain.Word
		for _, f := range strings.Fields(s) {
			out = append(out, domain.Word{Text: f, Literal: f != "$X"})
		}
		return out
	}
	tests := []struct {
		args string
		want []string
	}{
		{"build", []string{"build"}},
		{"build --release extra", []string{"build"}},
		{"build ::: lint:go --fix ::: fmt", []string{"build", "lint:go", "fmt"}},
		{"-j 4 --force build", []string{"build"}},
		{"--timeout 10m -- test", []string{"test"}},
		{"lint:*", nil},
		{"$X", nil},
		{"", nil},
		{"-C web build", nil},
		{"-Cweb build", nil},
		{"--cd=web build", nil},
	}
	m := &Mise{}
	for _, tt := range tests {
		// Act
		names := m.Names(words(tt.args))

		// Assert
		var got []string
		for _, w := range names {
			got = append(got, w.Text)
		}
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("Names(%q) (-want +got):\n%s", tt.args, diff)
		}
	}
}
