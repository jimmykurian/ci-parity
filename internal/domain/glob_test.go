// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

import "testing"

func TestMatchPath(t *testing.T) {
	tests := []struct {
		pattern, name string
		want          bool
	}{
		{".github/workflows/*.yml", ".github/workflows/ci.yml", true},
		{".github/workflows/*.yml", ".github/workflows/ci.yaml", false},
		{".github/workflows/*.yml", ".github/workflows/sub/ci.yml", false},
		{".github/workflows/?i.yml", ".github/workflows/ci.yml", true},
		{".github/workflows/?.yml", ".github/workflows/ci.yml", false},
		// ** in the middle: zero, one and several segments.
		{".github/actions/**/action.yml", ".github/actions/action.yml", true},
		{".github/actions/**/action.yml", ".github/actions/build/action.yml", true},
		{".github/actions/**/action.yml", ".github/actions/build/node/action.yml", true},
		{".github/actions/**/action.yml", ".github/actions/build/action.yaml", false},
		// ** at the start and at the end.
		{"**/action.yml", "action.yml", true},
		{"**/action.yml", "a/b/c/action.yml", true},
		{".github/**", ".github", true},
		{".github/**", ".github/workflows/ci.yml", true},
		{".github/**", "docs/ci.yml", false},
		// * inside a segment never crosses /.
		{".github/*", ".github/workflows/ci.yml", false},
		{".github/*/ci.yml", ".github/workflows/ci.yml", true},
		// Brackets and other syntax are literal.
		{".github/[ab].yml", ".github/[ab].yml", true},
		{".github/[ab].yml", ".github/a.yml", false},
	}
	for _, tt := range tests {
		if got := MatchPath(tt.pattern, tt.name); got != tt.want {
			t.Errorf("MatchPath(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
		}
	}
}

func TestMatchCommand(t *testing.T) {
	tests := []struct {
		pattern, command string
		want             bool
	}{
		{"gh release *", "gh release create v1.0.0", true},
		{"gh release *", "gh release upload dist/acme-app.tar.gz", true},
		{"gh release *", "gh release", false},
		{"gh release*", "gh release", true},
		{"docker build *", "docker push widget-service", false},
		{"npm ?i", "npm ci", true},
		{"npm ?i", "npm i", false},
		{"NPM ci", "npm ci", false},
		{"*", "", true},
		{"pnpm test", "pnpm test", true},
	}
	for _, tt := range tests {
		if got := MatchCommand(tt.pattern, tt.command); got != tt.want {
			t.Errorf("MatchCommand(%q, %q) = %v, want %v", tt.pattern, tt.command, got, tt.want)
		}
	}
}

func TestStaticPrefix(t *testing.T) {
	tests := map[string]string{
		".github/workflows/*.yml":       ".github/workflows",
		".github/actions/**/action.yml": ".github/actions",
		"**/action.yml":                 "",
		"action.yml":                    "",
	}
	for pattern, want := range tests {
		if got := StaticPrefix(pattern); got != want {
			t.Errorf("StaticPrefix(%q) = %q, want %q", pattern, got, want)
		}
	}
}
