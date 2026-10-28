package rules

import (
	"strings"
	"testing"
)

func TestClosest(t *testing.T) {
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
		if got := s.closest(unknown); got != want {
			t.Errorf("closest(%q) = %q, want %q", unknown, got, want)
		}
	}
}

func TestCanonical(t *testing.T) {
	tests := map[string]string{
		"  pnpm   test ":  "pnpm test",
		"pnpm run lint":   "pnpm lint",
		"yarn run build":  "yarn build",
		"npm run lint":    "npm run lint",
		"pnpm run":        "pnpm run",
		"go test   ./...": "go test ./...",
	}
	for in, want := range tests {
		if got := canonical(in); got != want {
			t.Errorf("canonical(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUnwrap(t *testing.T) {
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
		got := ""
		if rest, ok := unwrap(strings.Fields(in)); ok {
			got = strings.Join(rest, " ")
		}
		if got != want {
			t.Errorf("unwrap(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCommandSuggestionUnwraps(t *testing.T) {
	s := &suggester{byCommand: map[string]string{}}
	for name, cmd := range map[string]string{"lint": "eslint .", "test": "uv run pytest", "fmt": "npx prettier --check .", "check": "pnpm run check"} {
		for _, key := range keys(cmd) {
			s.byCommand[key] = name
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
		if got := s.command(in); got != want {
			t.Errorf("command(%q) = %q, want %q", in, got, want)
		}
	}
}
