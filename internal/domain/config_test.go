// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToolSetExtendsAndRemoves(t *testing.T) {
	// Arrange
	cfg := Default()
	cfg.Tools.Extend = []string{"k6x"}
	cfg.Tools.Remove = []string{"go"}

	// Act
	set := cfg.ToolSet()

	// Assert
	if !set["pnpm"] || !set["k6x"] || set["go"] {
		t.Errorf("extend/remove not applied: pnpm=%v k6x=%v go=%v", set["pnpm"], set["k6x"], set["go"])
	}
}

func TestToolSetReplaces(t *testing.T) {
	// Arrange
	cfg := Default()
	cfg.Tools.Replace = []string{"widgetctl"}
	want := map[string]bool{"widgetctl": true}

	// Act
	set := cfg.ToolSet()

	// Assert
	if diff := cmp.Diff(want, set); diff != "" {
		t.Errorf("replace (-want +got):\n%s", diff)
	}
}

func TestToolSetFollowsRunner(t *testing.T) {
	// Arrange
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
		// Act
		set := (&Config{Runner: tt.runner, Tools: tt.tools}).ToolSet()

		// Assert
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
