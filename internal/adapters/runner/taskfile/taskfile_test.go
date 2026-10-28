// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package taskfile

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/rules"
)

var _ rules.TaskSource = (*Taskfile)(nil)

// writeTree writes files (path -> content) under a new temp dir and returns it.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestLoadThreeLevelsAndCycle(t *testing.T) {
	root := writeTree(t, map[string]string{
		"Taskfile.yml": `version: "3"
includes:
  web: ./web
tasks:
  build: go build ./...
`,
		"web/Taskfile.yml": `version: "3"
includes:
  ui:
    taskfile: ./ui/Taskfile.yml
    aliases: [frontend]
tasks:
  lint:
    cmds: [pnpm lint]
`,
		"web/ui/Taskfile.yml": `version: "3"
includes:
  back: ../../Taskfile.yml
tasks:
  default:
    cmds: [pnpm dev]
  build:
    aliases: [b]
    cmds:
      - pnpm build
`,
	})
	tf, err := Load(root, "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"build",
		"web:lint",
		"web:frontend", "web:frontend:b", "web:frontend:build", "web:frontend:default",
		"web:ui", "web:ui:b", "web:ui:build", "web:ui:default",
	}
	if diff := cmp.Diff(sorted(want), tf.Tasks()); diff != "" {
		t.Errorf("Tasks (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"pnpm build"}, tf.Commands("web:ui:build")); diff != "" {
		t.Errorf("Commands(web:ui:build) (-want +got):\n%s", diff)
	}
	for _, name := range []string{"web:frontend:build", "web:ui:b", "web:ui"} {
		if got := tf.Commands(name); got != nil {
			t.Errorf("Commands(%q) = %v, want none for an alias or namespace shortcut", name, got)
		}
	}
	if tf.File() != "Taskfile.yml" {
		t.Errorf("File = %q", tf.File())
	}
}

func TestLoadIncludeOptions(t *testing.T) {
	root := writeTree(t, map[string]string{
		"Taskfile.yml": `version: '3'
includes:
  tools:
    taskfile: tools.yml
    internal: true
  shared:
    taskfile: shared.yml
    flatten: true
  extra:
    taskfile: missing.yml
    optional: true
tasks:
  ci:
    cmds:
      - task: lint
      - cmd: go vet ./...
      - cmd: echo {{.ITEM}}
        for: [a, b]
      - defer: rm -rf tmp
  setup:
    internal: true
    cmd: pnpm install
  list-form:
    - pnpm test
    - task: setup
`,
		"tools.yml":  "tasks:\n  fmt: gofmt -w .\n",
		"shared.yml": "version: 3\ntasks:\n  lint: golangci-lint run\n",
	})
	tf, err := Load(root, "Taskfile.yml")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"ci", "lint", "list-form", "setup", "tools:fmt"}, tf.Tasks()); diff != "" {
		t.Errorf("Tasks (-want +got):\n%s", diff)
	}
	internal := map[string]bool{"setup": true, "tools:fmt": true}
	for _, name := range tf.Tasks() {
		if tf.IsInternal(name) != internal[name] {
			t.Errorf("IsInternal(%q) = %v", name, tf.IsInternal(name))
		}
	}
	if diff := cmp.Diff([]string{"go vet ./..."}, tf.Commands("ci")); diff != "" {
		t.Errorf("Commands(ci) skips task, for and defer entries (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"pnpm test"}, tf.Commands("list-form")); diff != "" {
		t.Errorf("Commands(list-form) (-want +got):\n%s", diff)
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"no Taskfile", map[string]string{"README.md": "x"}, "no Taskfile found"},
		{"missing include", map[string]string{"Taskfile.yml": "includes:\n  api: ./api.yml\n"}, `include "api" in Taskfile.yml`},
		{"templated include", map[string]string{"Taskfile.yml": "includes:\n  os: ./{{OS}}.yml\n"}, "uses a template"},
		{"version 2", map[string]string{"Taskfile.yml": "version: '2'\n"}, `unsupported Taskfile version "2"`},
		{"version in include", map[string]string{
			"Taskfile.yml": "includes:\n  old: old.yml\n", "old.yml": "version: 2\n",
		}, "parse old.yml"},
		{"bad yaml", map[string]string{"Taskfile.yml": "tasks: [\n"}, "parse Taskfile.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeTree(t, tt.files), "")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Load error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestDiscoverOrder(t *testing.T) {
	root := writeTree(t, map[string]string{
		"Taskfile.dist.yml": "tasks:\n  dist: x\n",
		"taskfile.yaml":     "tasks:\n  lower: x\n",
	})
	tf, err := Load(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if tf.File() != "taskfile.yaml" {
		t.Errorf("File = %q, want taskfile.yaml before the .dist variants", tf.File())
	}
}

func sorted(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}
