package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseDefaults(t *testing.T) {
	for _, src := range []string{"", "# only a comment\n", "runner: taskfile\n", "workflows:\n"} {
		cfg, err := Parse([]byte(src))
		if err != nil {
			t.Fatalf("Parse(%q): %v", src, err)
		}
		if diff := cmp.Diff(Default(), cfg); diff != "" {
			t.Errorf("Parse(%q) differs from defaults (-want +got):\n%s", src, diff)
		}
	}
}

func TestParseOverrides(t *testing.T) {
	src := `
workflows: [".github/workflows/*.yml"]
invokers:
  shell: [task]
tools:
  extend: [k6x]
allow:
  - match: "docker push *"
    reason: images publish from CI only
    workflows: [".github/workflows/release.yml"]
`
	cfg, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	want := Default()
	want.Workflows = []string{".github/workflows/*.yml"}
	want.Invokers.Shell = []string{"task"}
	want.Tools.Extend = []string{"k6x"}
	want.Allow = []AllowRule{{
		Match:     "docker push *",
		Reason:    "images publish from CI only",
		Workflows: []string{".github/workflows/release.yml"},
	}}
	if diff := cmp.Diff(want, cfg); diff != "" {
		t.Errorf("Parse mismatch (-want +got):\n%s", diff)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name, src, want string
	}{
		{"unknown top-level key", "runnr: taskfile\n", `line 1: unknown key "runnr"`},
		{"unknown nested key", "tools:\n  extnd: [k6]\n", `line 2: unknown key "tools.extnd"`},
		{"unknown key in list item", "allow:\n  - match: x\n    reason: y\n    why: z\n", `line 4: unknown key "allow[0].why"`},
		{"list given a string", "workflows: ci.yml\n", "line 1: workflows must be a list"},
		{"string given a list", "taskfile: [a, b]\n", "line 1: taskfile must be a string"},
		{"list item given a mapping", "transparent:\n  - {a: b}\n", "line 2: transparent[0] must be a string"},
		{"top level not a mapping", "- a\n", "line 1: the config must be a mapping"},
		{"allow without reason", "allow:\n  - match: \"gh *\"\n", "allow[0].reason is required"},
		{"allow without match", "allow:\n  - reason: r\n", "allow[0].match is required"},
		{"replace with extend", "tools:\n  replace: [pnpm]\n  extend: [k6]\n", "tools.replace cannot be combined"},
		{"other runner", "runner: make\n", `runner: unsupported runner "make"`},
		{"action invoker without input", "invokers:\n  actions:\n    - uses: ./x\n", "invokers.actions[0] needs both uses and input"},
		{"bad yaml", "tools: [\n", "did not find expected node content"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.src))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Parse error = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	root := t.TempDir()
	cfg, err := Load(root, DefaultPath, false)
	if err != nil || cfg.Runner != "taskfile" {
		t.Fatalf("Load missing optional config = %v, %v; want defaults", cfg, err)
	}
	if _, err := Load(root, "custom.yml", true); err == nil {
		t.Error("Load missing required config: want error")
	}
	if err := os.WriteFile(filepath.Join(root, DefaultPath), []byte("bogus: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = Load(root, DefaultPath, false)
	if err == nil || !strings.Contains(err.Error(), `config .ci-parity.yml: line 1: unknown key "bogus"`) {
		t.Errorf("Load error = %v", err)
	}
}

func TestToolSet(t *testing.T) {
	cfg := Default()
	cfg.Tools.Extend = []string{"k6x"}
	cfg.Tools.Remove = []string{"go"}
	set := cfg.ToolSet()
	if !set["pnpm"] || !set["k6x"] || set["go"] {
		t.Errorf("extend/remove not applied: pnpm=%v k6x=%v go=%v", set["pnpm"], set["k6x"], set["go"])
	}
	cfg = Default()
	cfg.Tools.Replace = []string{"widgetctl"}
	if diff := cmp.Diff(map[string]bool{"widgetctl": true}, cfg.ToolSet()); diff != "" {
		t.Errorf("replace (-want +got):\n%s", diff)
	}
}
