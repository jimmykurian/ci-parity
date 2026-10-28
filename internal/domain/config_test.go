package domain

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func TestToolSetFollowsRunner(t *testing.T) {
	tests := []struct {
		runner  string
		tools   ToolList
		in, out []string
	}{
		{RunnerTaskfile, ToolList{}, []string{"make", "just", "mise"}, []string{"task", "go-task"}},
		{RunnerMake, ToolList{}, []string{"task", "go-task", "just", "mise"}, []string{"make"}},
		{RunnerMake, ToolList{Extend: []string{"make"}}, []string{"make"}, nil},
		{RunnerMake, ToolList{Replace: []string{"pnpm"}}, []string{"pnpm"}, []string{"task", "make"}},
	}
	for _, tt := range tests {
		set := (&Config{Runner: tt.runner, Tools: tt.tools}).ToolSet()
		for _, tool := range tt.in {
			if !set[tool] {
				t.Errorf("runner %s, %+v: %s is not a tool", tt.runner, tt.tools, tool)
			}
		}
		for _, tool := range tt.out {
			if set[tool] {
				t.Errorf("runner %s, %+v: %s is a tool", tt.runner, tt.tools, tool)
			}
		}
	}
}
