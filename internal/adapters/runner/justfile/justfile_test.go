package justfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
	"github.com/jimmykurian/ci-parity/internal/rules"
)

var _ rules.TaskSource = (*Justfile)(nil)

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

const main = `set shell := ["bash", "-c"]
version := "1.2.3"
export GOFLAGS := "-mod=mod"

# Build everything.
build:
    go build ./...

test *args: build
    @go test {{args}} ./...

deploy env region="eu-west-1:a":
    -./deploy.sh {{env}} {{region}}

[private]
release:
    #!/usr/bin/env bash
    goreleaser release

_helper:
    echo hidden

[group('lint'), private]
fmt:
    gofmt -l .

lint +paths:
    golangci-lint run {{paths}}

alias b := build
[private]
alias r := release

import 'mk/ci.just'
import? 'missing.just'
mod db
mod? absent
`

func load(t *testing.T) *Justfile {
	t.Helper()
	root := write(t, map[string]string{
		"justfile":          main,
		"mk/ci.just":        "ci:\n    just test\n",
		"db/mod.just":       "migrate target='latest':\n    goose up {{target}}\nmod seed 'seed/data.just'\n",
		"db/seed/data.just": "load:\n    psql -f seed.sql\n",
	})
	j, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func TestRecipes(t *testing.T) {
	j := load(t)
	want := []string{"_helper", "b", "build", "ci", "db::migrate", "db::seed::load", "deploy", "fmt", "lint", "r", "release", "test"}
	if diff := cmp.Diff(want, j.Tasks()); diff != "" {
		t.Errorf("recipes (-want +got):\n%s", diff)
	}
	for name, internal := range map[string]bool{"build": false, "release": true, "_helper": true, "fmt": true, "r": true, "b": false, "db::migrate": false} {
		if got := j.IsInternal(name); got != internal {
			t.Errorf("IsInternal(%s) = %v, want %v", name, got, internal)
		}
	}
	commands := map[string][]string{
		"build":       {"go build ./..."},
		"test":        {"go test {{args}} ./..."},
		"deploy":      {"./deploy.sh {{env}} {{region}}"},
		"release":     {"goreleaser release"},
		"db::migrate": {"goose up {{target}}"},
		"b":           nil,
	}
	for name, want := range commands {
		if diff := cmp.Diff(want, j.Commands(name)); diff != "" {
			t.Errorf("Commands(%s) (-want +got):\n%s", name, diff)
		}
	}
	if j.File() != "justfile" || j.Command() != "just" {
		t.Errorf("File=%q Command=%q", j.File(), j.Command())
	}
}

func TestMissingFiles(t *testing.T) {
	if _, err := Load(write(t, map[string]string{"README.md": "x"})); err == nil || !strings.Contains(err.Error(), "no justfile found") {
		t.Errorf("no justfile: err = %v", err)
	}
	if _, err := Load(write(t, map[string]string{"justfile": "import 'nope.just'\n"})); err == nil || !strings.Contains(err.Error(), `import "nope.just" not found`) {
		t.Errorf("missing import: err = %v", err)
	}
	if _, err := Load(write(t, map[string]string{"justfile": "mod nope\n"})); err == nil || !strings.Contains(err.Error(), `module "nope" not found`) {
		t.Errorf("missing module: err = %v", err)
	}
	if _, err := Load(write(t, map[string]string{"justfile": "import 'a.just'\n", "a.just": "import 'justfile'\nx:\n    true\n"})); err != nil {
		t.Errorf("import cycle: err = %v", err)
	}
}

func TestNames(t *testing.T) {
	j := load(t)
	words := func(s string) []domain.Word {
		var out []domain.Word
		for i, f := range strings.Fields(s) {
			out = append(out, domain.Word{Text: f, Literal: f != "$X", Line: 1, Col: i + 1})
		}
		return out
	}
	tests := []struct {
		args string
		want []string
	}{
		{"build", []string{"build"}},
		{"build test", []string{"build", "test"}},
		{"deploy prod build", []string{"deploy"}}, // region default takes "build"
		{"deploy prod", []string{"deploy"}},
		{"test -v -race", []string{"test"}}, // variadic takes the rest
		{"lint ./a ./b", []string{"lint"}},
		{"b ci", []string{"b", "ci"}},
		{"db migrate", []string{"db::migrate"}},
		{"db migrate v42 build", []string{"db::migrate", "build"}},
		{"db::migrate", []string{"db::migrate"}},
		{"db seed load", []string{"db::seed::load"}},
		{"db", nil}, // the module's default recipe
		{"version=2 build", []string{"build"}},
		{"--set version 2 build", []string{"build"}},
		{"--dotenv-path .env.ci --yes build", []string{"build"}},
		{"biuld extra words", []string{"biuld"}}, // stops after an unknown recipe
		{"$X build", nil},
		{"", nil},
		{"--list", nil},
		{"-f other.just build", nil},
		{"--justfile=other.just build", nil},
		{"-d web build", nil},
		{"--working-directory web build", nil},
		{"--show build", nil},
	}
	for _, tt := range tests {
		var got []string
		for _, w := range j.Names(words(tt.args)) {
			got = append(got, w.Text)
		}
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("Names(%q) (-want +got):\n%s", tt.args, diff)
		}
	}
}
