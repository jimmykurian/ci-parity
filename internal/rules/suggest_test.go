// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package rules

import (
	"strings"
	"testing"
)

func TestClosest(t *testing.T) {
	// Arrange
	s := &suggester{names: []string{"build", "ds", "lint", "lint:ci", "test:unit", "web:lint", "web:ui:build"}}
	tests := map[string]string{
		"biuld":        "build",        // distance 2
		"lnt":          "lint",         // distance 1
		"lint:all":     "lint:ci",      // distance 3
		"test":         "test:unit",    // prefix wins over "ds" at distance 3
		"web":          "web:lint",     // shortest prefix match
		"web:ui:bild":  "web:ui:build", // nested name
		"publish-docs": "",             // nothing within distance 3
		"xy":           "",             // distance 2 to "ds" is not less than the length
		"lintt":        "lint",         // distance 1 beats "lint:ci" at distance 3
	}
	for unknown, want := range tests {
		// Act
		got := s.closest(unknown)

		// Assert
		if got != want {
			t.Errorf("closest(%q) = %q, want %q", unknown, got, want)
		}
	}
}

func TestCanonical(t *testing.T) {
	// Arrange
	tests := map[string]string{
		"  pnpm   test ":  "pnpm test",
		"pnpm run lint":   "pnpm lint",
		"yarn run build":  "yarn build",
		"npm run lint":    "npm run lint",
		"pnpm run":        "pnpm run",
		"go test   ./...": "go test ./...",
	}
	for in, want := range tests {
		// Act
		got := canonical(in)

		// Assert
		if got != want {
			t.Errorf("canonical(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUnwrap(t *testing.T) {
	// Arrange
	tests := map[string]string{
		"npx eslint .":                             "eslint .",
		"pnpx eslint .":                            "eslint .",
		"pnx eslint .":                             "eslint .",
		"bunx eslint .":                            "eslint .",
		"npx --yes -p eslint@9 eslint .":           "eslint .",
		"npx --package=eslint@9 eslint .":          "eslint .",
		"npx -- eslint --fix":                      "eslint --fix",
		"pnpm exec tsc --noEmit":                   "tsc --noEmit",
		"pnpm dlx create-vite":                     "create-vite",
		"uv run pytest -q":                         "pytest -q",
		"uv run --with pytest-cov --frozen pytest": "pytest",
		"uvx --python 3.12 ruff check .":           "ruff check .",
		"dotnet tool run dotnet-format":            "dotnet-format",
		"npx -c 'eslint .'":                        "", // -c hides the command
		"npx --call=eslint":                        "",
		"npx --yes":                                "", // nothing left
		"pnpm test":                                "", // not a runner
		"pnpm exec":                                "",
		"uv pip install x":                         "",
	}
	for in, want := range tests {
		// Act
		rest, ok := unwrap(strings.Fields(in))

		// Assert
		got := ""
		if ok {
			got = strings.Join(rest, " ")
		}
		if got != want {
			t.Errorf("unwrap(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCommandSuggestionUnwraps(t *testing.T) {
	// Arrange
	s := &suggester{byCommand: map[string][]candidate{}, invoke: "task", reaches: func(string) bool { return true }}
	for name, cmd := range map[string]string{"lint": "eslint .", "test": "uv run pytest", "fmt": "npx prettier --check .", "check": "pnpm run check"} {
		for _, key := range keys(cmd) {
			s.byCommand[key] = append(s.byCommand[key], candidate{name: name, dirKnown: true})
		}
	}
	tests := map[string]string{
		"npx eslint .":       "task lint", // inline unwrapped
		"pytest":             "task test", // task unwrapped
		"uvx pytest":         "task test", // both unwrapped
		"prettier --check .": "task fmt",
		"eslint .":           "task lint",  // as written
		"pnpm check":         "task check", // run rule still applies
		"npx -c 'eslint .'":  "",
		"npx eslint src":     "",
	}
	for in, want := range tests {
		// Act
		got := s.command(in, "", true)

		// Assert
		if got != want {
			t.Errorf("command(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCommandSuggestionMatchesTheWorkingDirectory(t *testing.T) {
	// Arrange
	s := &suggester{invoke: "task", reaches: func(dir string) bool { return dir != "docs" }, byCommand: map[string][]candidate{
		"pnpm test": {
			{name: "test", dir: "", dirKnown: true},
			{name: "web:test", dir: "web", dirKnown: true},
			{name: "gen:test", dirKnown: false},
			{name: "docs:test", dir: "docs", dirKnown: true},
		},
		"pnpm build": {{name: "web:build", dir: "web", dirKnown: true}},
	}}
	tests := []struct {
		name, cmd, dir string
		dirKnown       bool
		want           string
	}{
		{"root step, root task", "pnpm test", "", true, "task test"},
		{"step in web, task in web", "pnpm test", "web", true, "task web:test"},
		{"step in web, only a root task would run elsewhere", "pnpm build", "", true, ""},
		{"step in another directory", "pnpm test", "api", true, ""},
		{"step directory unknown", "pnpm test", "", false, ""},
		{"the runner reads another file from the step's directory", "pnpm test", "docs", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := s.command(tt.cmd, tt.dir, tt.dirKnown)

			// Assert
			if got != tt.want {
				t.Errorf("command(%q, %q, %v) = %q, want %q", tt.cmd, tt.dir, tt.dirKnown, got, tt.want)
			}
		})
	}
}
