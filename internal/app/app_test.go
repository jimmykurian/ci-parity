package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

type fakeConfig struct{}

func (fakeConfig) Load(string) (*domain.Config, error) {
	cfg := domain.Default()
	cfg.Workflows = []string{"*.yml"}
	return cfg, nil
}

// fakeWorkflows maps each path to its number of bash run: steps.
type fakeWorkflows map[string]int

func (w fakeWorkflows) Discover([]string) ([]string, error) {
	return []string{"a.yml", "b.yml"}, nil
}

func (w fakeWorkflows) Load(path string) (*domain.File, error) {
	f := &domain.File{Path: path, Kind: domain.KindWorkflow}
	for i := range w[path] {
		f.Steps = append(f.Steps, domain.Step{
			File: path, Job: "build", Index: i + 1, Shell: "bash", Line: i + 1, Col: 1,
			Run: &domain.Script{Text: "pnpm test", Line: i + 1, Col: 6},
		})
	}
	return f, nil
}

type fakeTasks struct{}

func (fakeTasks) Load(string) (TaskSource, error) { return fakeTasks{}, nil }
func (fakeTasks) File() string                    { return "Taskfile.yml" }
func (fakeTasks) Tasks() []string                 { return nil }
func (fakeTasks) IsInternal(string) bool          { return false }
func (fakeTasks) Commands(string) []string        { return nil }

// fakeExtractor reports every script as one inline pnpm call.
type fakeExtractor struct{}

func (fakeExtractor) Supports(sh string) bool { return sh == "bash" }

func (fakeExtractor) Extract(script string, _ domain.ExtractOptions) (domain.Extraction, error) {
	return domain.Extraction{Commands: []domain.Command{
		{Kind: domain.KindTool, Name: "pnpm", Normalized: script, Line: 1, Col: 1},
	}}, nil
}

type fakeStore struct {
	baseline *domain.Baseline // nil means no file
	written  map[string]int
}

func (s *fakeStore) Read(path string) (*domain.Baseline, error) {
	if s.baseline == nil {
		return nil, fmt.Errorf("read baseline %s: %w", path, fs.ErrNotExist)
	}
	return s.baseline, nil
}

func (s *fakeStore) Write(_ string, counts map[string]int) error {
	s.written = counts
	return nil
}

type fakeReporter struct {
	findings []domain.Finding
	rows     []domain.Row
}

func (r *fakeReporter) Report(findings []domain.Finding, rows []domain.Row) error {
	r.findings, r.rows = findings, rows
	return nil
}

// newTestApp scans a.yml with two inline calls and b.yml with one.
func newTestApp(b *domain.Baseline) (*App, *fakeStore) {
	store := &fakeStore{baseline: b}
	return &App{
		Config:    fakeConfig{},
		Workflows: fakeWorkflows{"a.yml": 2, "b.yml": 1},
		Tasks:     fakeTasks{},
		Extractor: fakeExtractor{},
		Baselines: store,
	}, store
}

func severities(findings []domain.Finding) []string {
	var out []string
	for _, f := range findings {
		out = append(out, f.File+":"+f.Rule+":"+string(f.Severity))
	}
	return out
}

func TestCheck(t *testing.T) {
	tolerant := &domain.Baseline{Workflows: map[string]int{"a.yml": 2, "b.yml": 1}}
	tests := []struct {
		name     string
		baseline *domain.Baseline
		strict   bool
		want     []string
		wantRows int
		failed   bool
	}{
		{
			"no baseline file is strict", nil, false,
			[]string{"a.yml:P001:error", "a.yml:P001:error", "b.yml:P001:error"},
			0, true,
		},
		{
			"baseline tolerates", tolerant, false,
			[]string{"a.yml:P001:tolerated", "a.yml:P001:tolerated", "b.yml:P001:tolerated"},
			2, false,
		},
		{
			"strict ignores the baseline", tolerant, true,
			[]string{"a.yml:P001:error", "a.yml:P001:error", "b.yml:P001:error"},
			0, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := newTestApp(tt.baseline)
			rep := &fakeReporter{}
			res, err := a.Check(context.Background(), CheckOptions{Strict: tt.strict, Reporter: rep})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, severities(res.Findings)); diff != "" {
				t.Errorf("findings (-want +got):\n%s", diff)
			}
			if len(res.Rows) != tt.wantRows || res.Failed() != tt.failed {
				t.Errorf("rows = %d, failed = %v; want %d, %v", len(res.Rows), res.Failed(), tt.wantRows, tt.failed)
			}
			if diff := cmp.Diff(res.Findings, rep.findings); diff != "" {
				t.Errorf("reported findings differ from the result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriteBaseline(t *testing.T) {
	lower := &domain.Baseline{Workflows: map[string]int{"a.yml": 1, "b.yml": 1}}
	tests := []struct {
		name     string
		baseline *domain.Baseline
		allow    bool
		wantRegs []domain.Row
	}{
		{"first run never refuses", nil, false, nil},
		{"refuses a raised count", lower, false, []domain.Row{{File: "a.yml", Baseline: 1, Current: 2}}},
		{"allow-regression writes", lower, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, store := newTestApp(tt.baseline)
			res, err := a.WriteBaseline(context.Background(), BaselineOptions{AllowRegression: tt.allow})
			var regs *RegressionError
			if tt.wantRegs != nil {
				if !errors.As(err, &regs) {
					t.Fatalf("err = %v, want *RegressionError", err)
				}
				if diff := cmp.Diff(tt.wantRegs, regs.Rows); diff != "" {
					t.Errorf("regressions (-want +got):\n%s", diff)
				}
				if store.written != nil {
					t.Errorf("wrote %v after refusing", store.written)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := BaselineResult{Path: ".ci-parity-baseline.json", Workflows: 2, Findings: 3}
			if diff := cmp.Diff(want, res); diff != "" {
				t.Errorf("result (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(map[string]int{"a.yml": 2, "b.yml": 1}, store.written); diff != "" {
				t.Errorf("written (-want +got):\n%s", diff)
			}
		})
	}
}
