// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const sha = "0123456789abcdef0123456789abcdef01234567"

func TestParseReposReadsTheTrialList(t *testing.T) {
	// Arrange
	data, err := os.ReadFile("../../test/trial/repos.yml")
	if err != nil {
		t.Fatal(err)
	}

	// Act
	repos, err := ParseRepos(data)
	if err != nil {
		t.Fatalf("ParseRepos(repos.yml): %v", err)
	}

	// Assert
	if len(repos) != 10 {
		t.Errorf("repos = %d, want 10", len(repos))
	}
}

func TestParseReposRejectsBadEntries(t *testing.T) {
	// Arrange
	tests := []struct {
		name, yaml, want string
	}{
		{"empty list", "repos: []", "no repos"},
		{"not owner/name", "repos: [{repo: acme-app, commit: " + sha + ", runner: make}]", "not owner/name"},
		{"option in the name", "repos: [{repo: --upload-pack/x, commit: " + sha + ", runner: make}]", "not owner/name"},
		{"short commit", "repos: [{repo: acme/acme-app, commit: 0123456, runner: make}]", "not a full SHA"},
		{"unknown runner", "repos: [{repo: acme/acme-app, commit: " + sha + ", runner: bazel}]", "unknown runner"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := ParseRepos([]byte(tt.yaml))

			// Assert
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseRepos() error = %v, want one containing %q", err, tt.want)
			}
		})
	}
}

func TestRepoDirJoinsOwnerAndName(t *testing.T) {
	// Act
	got := Repo{Repo: "acme/widget-service"}.Dir()

	// Assert
	if got != "acme_widget-service" {
		t.Errorf("Dir() = %q, want acme_widget-service", got)
	}
}

func TestParseResultCountsFindingsByRule(t *testing.T) {
	// Arrange
	out := `{"version":1,"summary":{"errors":3,"warnings":1,"tolerated":0,"allowed":0},"findings":[
		{"rule":"P001","file":"ci.yml","line":9},{"rule":"P001","file":"ci.yml","line":11},
		{"rule":"P002","file":"ci.yml","line":13},{"rule":"P005","file":"ci.yml","line":20}]}`
	want := map[string]int{"P001": 2, "P002": 1, "P005": 1}

	// Act
	res, err := ParseResult(Repo{Repo: "acme/acme-app"}, []byte(out))
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if diff := cmp.Diff(want, res.Count()); diff != "" {
		t.Errorf("Count() (-want +got):\n%s", diff)
	}
	if res.Errors != 3 || res.Warnings != 1 {
		t.Errorf("errors, warnings = %d, %d, want 3, 1", res.Errors, res.Warnings)
	}
}

func TestRenderTotalsAndListsTheHandCheckedFindings(t *testing.T) {
	// Arrange
	repo := Repo{Repo: "acme/acme-app", Commit: sha, Runner: "taskfile"}
	results := []Result{{
		Repo:   repo,
		Errors: 2,
		Findings: []Finding{
			{Rule: "P001", File: "ci.yml", Line: 9, Command: "pnpm test", Suggestion: "task test"},
			{Rule: "P002", File: "ci.yml", Line: 13, Message: "'task lint' is not defined in Taskfile.yml", Suggestion: "test"},
		},
	}}

	// Act
	got := Render(results)

	// Assert
	for _, want := range []string{
		"| acme/acme-app | 0123456 | taskfile | 1 | 1 | 0 | 0 | 0 | 2 | 0 |",
		"| **Total** | | | **1** | **1** | **0** | **0** | **0** | **2** | **0** |",
		"- acme/acme-app `ci.yml:13`: 'task lint' is not defined in Taskfile.yml",
		"- acme/acme-app `ci.yml:9`: `pnpm test` -> `task test`",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Render() has no line %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "-> `test`") {
		t.Errorf("Render() lists a P002's closest name as a suggestion:\n%s", got)
	}
}

func TestRenderSaysNoneForEmptySections(t *testing.T) {
	// Act
	got := Render([]Result{{Repo: Repo{Repo: "acme/acme-app", Commit: sha, Runner: "make"}}})

	// Assert
	if strings.Count(got, "None.") != 2 {
		t.Errorf("Render() should say None. under both sections:\n%s", got)
	}
}
