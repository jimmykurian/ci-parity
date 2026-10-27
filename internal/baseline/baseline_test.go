package baseline

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/rules"
)

func p001(file string, line int, sev rules.Severity) rules.Finding {
	return rules.Finding{Rule: rules.InlineTool, Severity: sev, File: file, Line: line, Col: 1}
}

func TestApply(t *testing.T) {
	b := &File{Version: 1, Workflows: map[string]int{"a.yml": 2, "b.yml": 0, "d.yml": 4}}
	findings := []rules.Finding{
		p001("a.yml", 3, rules.SeverityAllowed),
		p001("a.yml", 5, rules.SeverityError),
		p001("a.yml", 7, rules.SeverityError),
		p001("a.yml", 9, rules.SeverityError),
		p001("c.yml", 2, rules.SeverityError),
		{Rule: rules.Unsupported, Severity: rules.SeverityWarning, File: "a.yml", Line: 4},
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
	old := &File{Version: 1, Workflows: map[string]int{"a.yml": 2, "b.yml": 1}}
	got := Regressions(old, map[string]int{"a.yml": 3, "b.yml": 0, "new.yml": 1})
	want := []Row{{"a.yml", 2, 3}, {"new.yml", 0, 1}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Regressions (-want +got):\n%s", diff)
	}
}

func TestReadErrors(t *testing.T) {
	root := t.TempDir()
	if _, err := Read(root, "missing.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing file error = %v, want fs.ErrNotExist", err)
	}
	tests := map[string]string{
		`{"version": 2, "workflows": {}}`:             "unsupported version 2",
		`{"version": 1, "workflows": {"a.yml": -1}}`:  "negative count",
		`{"version": 1, "workflows": {}, "extra": 1}`: "unknown field",
		`{"version": 1, "workflows": {"a.yml": "x"}}`: "cannot unmarshal",
	}
	for content, want := range tests {
		if err := os.WriteFile(filepath.Join(root, "b.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Read(root, "b.json"); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Read(%s) error = %v, want %q", content, err, want)
		}
	}
}
