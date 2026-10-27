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
