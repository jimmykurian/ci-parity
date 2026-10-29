// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package makefile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
	"github.com/jimmykurian/ci-parity/internal/rules"
)

var _ rules.TaskSource = (*Makefile)(nil)

func write(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestTargetsAndCommands(t *testing.T) {
	// Arrange
	root := write(t, map[string]string{"Makefile": strings.Join([]string{
		"# A comment: not a target",
		"GO ?= go",
		"VERSION := 1.2.3",
		"FLAGS = -v -run=Test:x",
		"export CGO_ENABLED := 0",
		".PHONY: build test lint",
		".DEFAULT_GOAL := build",
		"",
		"build: deps",
		"\t@$(GO) build ./...",
		"",
		"\t-echo done",
		"test lint: build",
		"\tgo test ./...",
		"%.o: %.c",
		"\tcc -c $<",
		"$(BIN): main.go",
		"\tgo build -o $@",
		"release: VERSION = 2.0.0",
		"ci:: ; go vet ./...",
		"define HELP",
		"fake: target inside define",
		"endef",
		"long: \\",
		"  build",
		"\t+go run \\",
		"\t  ./tools/gen",
		"ifeq ($(CI),true)",
		"deps:",
		"\tgo mod download",
		"else",
		"\t# commented recipe line",
		"endif",
		"",
	}, "\n")})
	want := map[string][]string{
		"build": {"$(GO) build ./...", "echo done"},
		"test":  {"go test ./..."},
		"lint":  {"go test ./..."},
		"ci":    {"go vet ./..."},
		"long":  {"go run ./tools/gen"},
		"deps":  {"go mod download"},
	}

	// Act
	m, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if diff := cmp.Diff([]string{"build", "ci", "deps", "lint", "long", "test"}, m.Tasks()); diff != "" {
		t.Errorf("targets (-want +got):\n%s", diff)
	}
	for name, cmds := range want {
		if diff := cmp.Diff(cmds, m.Commands(name)); diff != "" {
			t.Errorf("commands of %s (-want +got):\n%s", name, diff)
		}
	}
	if m.File() != "Makefile" || m.Command() != "make" || m.IsInternal("build") {
		t.Errorf("File=%q Command=%q IsInternal=%v", m.File(), m.Command(), m.IsInternal("build"))
	}
}

func TestDiscoveryOrderAndIncludes(t *testing.T) {
	// Arrange
	root := write(t, map[string]string{
		"GNUmakefile": "include mk/ci.mk missing.mk $(GEN).mk\nall:\n\ttrue\n",
		"Makefile":    "ignored:\n\ttrue\n",
		"mk/ci.mk":    "-include mk/loop.mk\nci-lint:\n\tgolangci-lint run\n",
		"mk/loop.mk":  "include mk/ci.mk\nloop:\n\ttrue\n",
	})

	// Act
	m, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if m.File() != "GNUmakefile" {
		t.Errorf("File = %q, want GNUmakefile", m.File())
	}
	if diff := cmp.Diff([]string{"all", "ci-lint", "loop"}, m.Tasks()); diff != "" {
		t.Errorf("targets (-want +got):\n%s", diff)
	}
}

func TestMissingMakefile(t *testing.T) {
	// Arrange
	root := write(t, map[string]string{"README.md": "x"})

	// Act
	_, err := Load(root)

	// Assert
	if err == nil || !strings.Contains(err.Error(), "no Makefile found") {
		t.Errorf("err = %v, want no Makefile found", err)
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
		{"build lint", []string{"build", "lint"}},
		{"VERSION=1.2.3 build", []string{"build"}},
		{"-j 4 test", []string{"test"}},
		{"-j test", []string{"test"}},
		{"--jobs=8 -k test", []string{"test"}},
		{"-I mk -W main.c build", []string{"build"}},
		{"-s -- -weird", []string{"-weird"}},
		{"$X build", []string{"build"}},
		{"", nil},
		{"-j4", nil},
		{"-C web build", nil},
		{"-Cweb build", nil},
		{"--directory=web build", nil},
		{"-f other.mk build", nil},
		{"--file=other.mk build", nil},
		{"--makefile other.mk build", nil},
	}
	m := &Makefile{}
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

func TestRunsMore(t *testing.T) {
	// Arrange
	root := write(t, map[string]string{"Makefile": strings.Join([]string{
		".PHONY: build vet",
		"build:",
		"\tgo build ./...",
		"vet: build",
		"\tgo vet ./...",
		"test: | tmp",
		"\tgo test ./...",
		"lint: ; golangci-lint run",
		"gen:",
		"gen: tools",
		"\tgo generate ./...",
		"",
	}, "\n")})
	want := map[string]bool{"build": false, "lint": false, "vet": true, "test": true, "gen": true}

	// Act
	m, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	for name, more := range want {
		if got := m.RunsMore(name); got != more {
			t.Errorf("RunsMore(%s) = %v, want %v", name, got, more)
		}
	}
}
