package rules

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/shell"
)

func TestTaskNames(t *testing.T) {
	tests := []struct {
		args string
		want []string
	}{
		{"build lint", []string{"build", "lint"}},
		{"build VERSION=1.2.3", []string{"build"}},
		{"-C 2 build", []string{"build"}},
		{"--concurrency=2 test", []string{"test"}},
		{"-o group --output prefixed build", []string{"build"}},
		{"--interval 5s lint", []string{"lint"}},
		{"--silent --force test", []string{"test"}},
		{"build -- --verbose extra", []string{"build"}},
		{"--list", nil},
		{"-t other.yml migrate", nil},
		{"--taskfile=other.yml migrate", nil},
		{"-d web build", nil},
		{"--dir=web build", nil},
		{"-dweb build", nil},
		{"--global setup", nil},
		{"-g setup", nil},
	}
	for _, tt := range tests {
		var args []shell.Word
		for _, f := range strings.Fields(tt.args) {
			args = append(args, shell.Word{Text: f, Literal: true})
		}
		var got []string
		for _, w := range taskNames(args) {
			got = append(got, w.Text)
		}
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("taskNames(%q) (-want +got):\n%s", tt.args, diff)
		}
	}
	dynamic := []shell.Word{{Text: "${{ inputs.target }}"}, {Text: "build", Literal: true}}
	if got := taskNames(dynamic); len(got) != 1 || got[0].Text != "build" {
		t.Errorf("non-literal names are skipped, got %v", got)
	}
}
