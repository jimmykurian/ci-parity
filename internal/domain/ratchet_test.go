// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func p001(file string, line int, sev Severity) Finding {
	return Finding{Rule: InlineTool, Severity: sev, File: file, Line: line, Col: 1}
}

func TestApply(t *testing.T) {
	b := &Baseline{Workflows: map[string]int{"a.yml": 2, "b.yml": 0, "d.yml": 4}}
	findings := []Finding{
		p001("a.yml", 3, SeverityAllowed),
		p001("a.yml", 5, SeverityError),
		p001("a.yml", 7, SeverityError),
		p001("a.yml", 9, SeverityError),
		p001("c.yml", 2, SeverityError),
		{Rule: Unsupported, Severity: SeverityWarning, File: "a.yml", Line: 4},
	}
	got, rows := Apply(b, []string{"a.yml", "b.yml", "c.yml", "d.yml"}, findings)

	var sev []string
	for _, f := range got {
		sev = append(sev, f.File+":"+f.Rule+":"+string(f.Severity))
	}
	want := []string{
		"a.yml:P004:error", "a.yml:P001:allowed", "a.yml:P005:warning",
		"a.yml:P001:tolerated", "a.yml:P001:tolerated", "a.yml:P001:error",
		"c.yml:P004:error", "c.yml:P001:error",
	}
	if diff := cmp.Diff(want, sev); diff != "" {
		t.Errorf("severities (-want +got):\n%s", diff)
	}
	wantRows := []Row{{"a.yml", 2, 3}, {"b.yml", 0, 0}, {"c.yml", 0, 1}, {"d.yml", 4, 0}}
	if diff := cmp.Diff(wantRows, rows); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
	if !strings.Contains(got[6].Message, "c.yml is locked at 0 and has 1 inline-tool findings") {
		t.Errorf("locked message = %q", got[6].Message)
	}
}

func TestRegressions(t *testing.T) {
	old := &Baseline{Workflows: map[string]int{"a.yml": 2, "b.yml": 1}}
	got := Regressions(old, map[string]int{"a.yml": 3, "b.yml": 0, "new.yml": 1})
	want := []Row{{"a.yml", 2, 3}, {"new.yml", 0, 1}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Regressions (-want +got):\n%s", diff)
	}
}
