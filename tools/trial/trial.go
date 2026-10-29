// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Rules are the report's rule columns, in order.
var Rules = []string{"P001", "P002", "P003", "P004", "P005"}

// Repo is one trial repository, pinned to a commit.
type Repo struct {
	Repo   string `yaml:"repo"`
	Commit string `yaml:"commit"`
	Runner string `yaml:"runner"`
}

// Dir is the repository's directory name in the cache: owner_name.
func (r Repo) Dir() string {
	return strings.ReplaceAll(r.Repo, "/", "_")
}

var (
	repoName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9][A-Za-z0-9_.-]*$`)
	fullSHA  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	runners  = map[string]bool{"taskfile": true, "make": true, "just": true, "mise": true}
)

// ParseRepos reads the trial list and checks each entry: an owner/name, a full
// commit SHA and a runner ci-parity supports.
func ParseRepos(data []byte) ([]Repo, error) {
	var list struct {
		Repos []Repo `yaml:"repos"`
	}
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("read trial list: %w", err)
	}
	if len(list.Repos) == 0 {
		return nil, fmt.Errorf("trial list has no repos")
	}
	for i, r := range list.Repos {
		switch {
		case !repoName.MatchString(r.Repo):
			return nil, fmt.Errorf("repo %d: %q is not owner/name", i+1, r.Repo)
		case !fullSHA.MatchString(r.Commit):
			return nil, fmt.Errorf("repo %s: commit %q is not a full SHA", r.Repo, r.Commit)
		case !runners[r.Runner]:
			return nil, fmt.Errorf("repo %s: unknown runner %q", r.Repo, r.Runner)
		}
	}
	return list.Repos, nil
}

// Finding is the part of a JSON-format finding the report uses.
type Finding struct {
	Rule       string `json:"rule"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Message    string `json:"message"`
	Command    string `json:"command"`
	Suggestion string `json:"suggestion"`
}

// Result is one repository's run: its findings and the summary counts.
type Result struct {
	Repo     Repo
	Findings []Finding
	Errors   int
	Warnings int
}

// ParseResult reads `ci-parity check --format json` output for repo.
func ParseResult(repo Repo, data []byte) (Result, error) {
	var out struct {
		Findings []Finding `json:"findings"`
		Summary  struct {
			Errors   int `json:"errors"`
			Warnings int `json:"warnings"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return Result{}, fmt.Errorf("read %s output: %w", repo.Repo, err)
	}
	return Result{Repo: repo, Findings: out.Findings, Errors: out.Summary.Errors, Warnings: out.Summary.Warnings}, nil
}

// Count returns how many findings of each rule the result has.
func (r Result) Count() map[string]int {
	n := map[string]int{}
	for _, f := range r.Findings {
		n[f.Rule]++
	}
	return n
}

// Render writes the report: a table of findings per repository and rule, then
// every P002 and every P001 suggestion, the findings checked by hand in the
// trial. (A P002's suggestion is the closest name, already in its message.)
func Render(results []Result) string {
	var b strings.Builder
	b.WriteString("| Repository | Commit | Runner | " + strings.Join(Rules, " | ") + " | Errors | Warnings |\n")
	b.WriteString("| --- | --- | --- |" + strings.Repeat(" ---: |", len(Rules)+2) + "\n")
	total := map[string]int{}
	errs, warns := 0, 0
	for _, r := range results {
		n := r.Count()
		fmt.Fprintf(&b, "| %s | %s | %s |", r.Repo.Repo, r.Repo.Commit[:7], r.Repo.Runner)
		for _, rule := range Rules {
			fmt.Fprintf(&b, " %d |", n[rule])
			total[rule] += n[rule]
		}
		fmt.Fprintf(&b, " %d | %d |\n", r.Errors, r.Warnings)
		errs += r.Errors
		warns += r.Warnings
	}
	b.WriteString("| **Total** | | |")
	for _, rule := range Rules {
		fmt.Fprintf(&b, " **%d** |", total[rule])
	}
	fmt.Fprintf(&b, " **%d** | **%d** |\n", errs, warns)

	section(&b, "P002 findings", results, func(f Finding) string {
		if f.Rule != "P002" {
			return ""
		}
		return f.Message
	})
	section(&b, "Suggestions", results, func(f Finding) string {
		if f.Rule != "P001" || f.Suggestion == "" {
			return ""
		}
		return fmt.Sprintf("`%s` -> `%s`", f.Command, f.Suggestion)
	})
	return b.String()
}

// section writes a heading and one line per finding that line returns text
// for, or "None." when there are none.
func section(b *strings.Builder, title string, results []Result, line func(Finding) string) {
	b.WriteString("\n## " + title + "\n\n")
	empty := true
	for _, r := range results {
		for _, f := range r.Findings {
			if text := line(f); text != "" {
				fmt.Fprintf(b, "- %s `%s:%d`: %s\n", r.Repo.Repo, f.File, f.Line, text)
				empty = false
			}
		}
	}
	if empty {
		b.WriteString("None.\n")
	}
}
