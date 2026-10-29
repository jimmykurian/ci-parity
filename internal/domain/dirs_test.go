// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"path/filepath"
	"testing"
)

func TestRepoDir(t *testing.T) {
	// Arrange
	tests := []struct {
		in      string
		want    string
		wantOK  bool
		comment string
	}{
		{"", "", true, "unset is the root"},
		{".", "", true, ""},
		{"./", "", true, ""},
		{"web", "web", true, ""},
		{"./web/", "web", true, ""},
		{"web/../api", "api", true, ""},
		{`web\e2e`, "web/e2e", true, "Windows separators"},
		{"${{ inputs.dir }}", "", false, "expression"},
		{"$HOME/web", "", false, "variable"},
		{"{{.DIR}}", "", false, "template"},
		{"~/web", "", false, "home directory"},
		{"/home/runner/work", "", false, "absolute"},
		{`C:\work`, "", false, "drive letter"},
		{"..", "", false, "outside the repository"},
		{"../other", "", false, "outside the repository"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			// Act
			got, ok := RepoDir(tt.in)

			// Assert
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("RepoDir(%q) = %q, %v, want %q, %v %s", tt.in, got, ok, tt.want, tt.wantOK, tt.comment)
			}
		})
	}
}

func TestReachesStopsAtANearerFile(t *testing.T) {
	// Arrange
	file := filepath.Join("repo", "Taskfile.yml")
	nearer := map[string]bool{filepath.Join("repo", "docs"): true}
	found := func(dir string) (string, bool) { return dir, nearer[dir] }
	tests := []struct {
		name, start string
		want        bool
	}{
		{"the file's own directory", "repo", true},
		{"below it, no file between", filepath.Join("repo", "web", "e2e"), true},
		{"a directory with its own file", filepath.Join("repo", "docs"), false},
		{"below a directory with its own file", filepath.Join("repo", "docs", "api"), false},
		{"outside the file's directory", "other", false},
		{"above the file's directory", ".", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := Reaches(file, tt.start, found)

			// Assert
			if got != tt.want {
				t.Errorf("Reaches(%q, %q) = %v, want %v", file, tt.start, got, tt.want)
			}
		})
	}
}
