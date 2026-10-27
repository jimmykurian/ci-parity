package rules

import "testing"

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
