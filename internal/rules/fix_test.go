package rules

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

func TestFixEdit(t *testing.T) {
	plain := func(v string) *domain.Script {
		return &domain.Script{Text: v, Value: v, SingleLine: true, Line: 5, Col: 14}
	}
	// cmd builds the command the extractor reports: its byte column in the value
	// and its normalized words.
	cmd := func(value, command string) domain.Command {
		return domain.Command{Line: 1, Col: strings.Index(value, command) + 1, Normalized: command}
	}
	tests := []struct {
		name    string
		run     *domain.Script
		command string
		want    *domain.Edit
	}{
		{"whole value", plain("pnpm test"), "pnpm test", &domain.Edit{File: "ci.yml", Line: 5, Col: 14, Old: "pnpm test", New: "task test"}},
		{"wrapper kept", plain("sudo -u ci pnpm test"), "pnpm test", &domain.Edit{File: "ci.yml", Line: 5, Col: 25, Old: "pnpm test", New: "task test"}},
		{"assignment kept", plain("CI=1 pnpm test"), "pnpm test", &domain.Edit{File: "ci.yml", Line: 5, Col: 19, Old: "pnpm test", New: "task test"}},
		{"characters, not bytes", plain("LANG=é pnpm test"), "pnpm test", &domain.Edit{File: "ci.yml", Line: 5, Col: 21, Old: "pnpm test", New: "task test"}},
		{"one-line block", &domain.Script{Value: "pnpm test\n", SingleLine: true, Block: true, Line: 9, Indent: 10}, "pnpm test", &domain.Edit{File: "ci.yml", Line: 9, Col: 11, Old: "pnpm test", New: "task test"}},
		{"chain", plain("pnpm build && pnpm test"), "pnpm test", nil},
		{"redirect", plain("pnpm test > out"), "pnpm test", nil},
		{"subshell prefix", plain("(cd web; pnpm test"), "pnpm test", nil},
		{"expansion prefix", plain("$RUNNER pnpm test"), "pnpm test", nil},
		{"expression", plain("pnpm test ${{ x }}"), "pnpm test", nil},
		{"not one line", &domain.Script{Value: "pnpm test", Line: 5, Col: 14}, "pnpm test", nil},
	}
	for _, tt := range tests {
		got := fixEdit("ci.yml", tt.run, cmd(strings.TrimSuffix(tt.run.Value, "\n"), tt.command), "task test")
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("%s (-want +got):\n%s", tt.name, diff)
		}
	}
}
