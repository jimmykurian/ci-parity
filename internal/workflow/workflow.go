// Package workflow loads GitHub Actions workflows and local composite actions
// and enumerates their steps with source positions.
package workflow

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Kind tells a workflow from a composite action.
type Kind int

const (
	// KindWorkflow is a file under .github/workflows with jobs.
	KindWorkflow Kind = iota + 1
	// KindAction is a local action.yml; only composite actions have steps.
	KindAction
)

// File is one scanned workflow or action.
type File struct {
	Path  string // repo-relative, forward slashes
	Kind  Kind
	Steps []Step
}

// Step is one job step or composite action step.
type Step struct {
	File  string
	Job   string // job ID; empty for composite actions
	Index int    // 1-based position within the job or action
	Name  string // the step's name:, empty when absent
	Shell string // resolved shell; empty for a composite step without shell
	Line  int    // position of the run: or uses: key
	Col   int
	Run   *Script
	Uses  string
	With  map[string]string
}

// Script is a step's run: value and where it sits in the file.
type Script struct {
	Text   string
	line   int       // source line of the script's first line
	col    int       // source column of the script's first character
	indent int       // indentation of block scalar content
	block  bool      // literal or folded block scalar
	segs   []segment // source positions of folded text; nil when not folded
}

// Position maps a 1-based line and column inside the script to a position in
// the file.
func (s *Script) Position(line, col int) (int, int) {
	switch {
	case s.segs != nil:
		return s.foldedPosition(line, col)
	case s.block:
		return s.line + line - 1, s.indent + col
	case line == 1:
		return s.line, s.col + col - 1
	default:
		return s.line + line - 1, col
	}
}

// Load reads and parses the workflow or action at rel under root.
func Load(root, rel string) (*File, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", rel, err)
	}
	var doc yaml.Node
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&doc); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parse %s: %w", rel, err)
	}
	f := &File{Path: rel, Kind: KindWorkflow}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return f, nil
	}
	top := doc.Content[0]
	lines := strings.Split(string(data), "\n")
	if runs := get(top, "runs"); runs != nil {
		f.Kind = KindAction
		if strings.EqualFold(scalar(get(runs, "using")), "composite") {
			f.Steps = steps(rel, "", get(runs, "steps"), "", lines)
		}
		return f, nil
	}
	defaultShell := shellDefault(top)
	jobs := get(top, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return f, nil
	}
	for i := 0; i+1 < len(jobs.Content); i += 2 {
		id, job := jobs.Content[i].Value, jobs.Content[i+1]
		if job.Kind != yaml.MappingNode {
			continue
		}
		shell := shellDefault(job)
		if shell == "" {
			shell = defaultShell
		}
		if shell == "" {
			shell = osDefaultShell(get(job, "runs-on"))
		}
		f.Steps = append(f.Steps, steps(rel, id, get(job, "steps"), shell, lines)...)
	}
	return f, nil
}

// steps enumerates a steps: sequence. fallback is the shell for steps that do
// not set one; it is empty for composite actions, where shell is required.
func steps(rel, job string, seq *yaml.Node, fallback string, lines []string) []Step {
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return nil
	}
	var out []Step
	for i, n := range seq.Content {
		if n.Kind != yaml.MappingNode {
			continue
		}
		s := Step{File: rel, Job: job, Index: i + 1, Name: scalar(get(n, "name")), Shell: fallback}
		if sh := scalar(get(n, "shell")); sh != "" {
			s.Shell = sh
		}
		if key, val := getPair(n, "run"); val != nil && val.Kind == yaml.ScalarNode {
			s.Line, s.Col = key.Line, key.Column
			s.Run = script(val, lines)
		} else if key, val := getPair(n, "uses"); val != nil {
			s.Line, s.Col = key.Line, key.Column
			s.Uses = val.Value
			s.With = stringMap(get(n, "with"))
		}
		out = append(out, s)
	}
	return out
}

func script(n *yaml.Node, lines []string) *Script {
	s := &Script{Text: n.Value, line: n.Line, col: n.Column}
	switch n.Style {
	case yaml.LiteralStyle:
		s.block = true
		s.line = n.Line + 1
		s.indent = blockIndent(lines, n.Line)
	case yaml.FoldedStyle:
		s.block = true
		s.line = n.Line + 1
		s.indent = blockIndent(lines, n.Line)
		s.segs = foldedBlock(lines, n.Line, s.indent, n.Value)
	case yaml.DoubleQuotedStyle, yaml.SingleQuotedStyle:
		s.col = n.Column + 1
	default:
		s.segs = foldedPlain(lines, n.Line, n.Column, n.Value)
	}
	// On a single-line value, `run: cmd # ci-parity: allow ...` ends in a YAML
	// comment, not a shell one; put it back so allow comments work there too.
	if !s.block && n.LineComment != "" && !strings.Contains(n.Value, "\n") {
		s.Text += " " + n.LineComment
	}
	return s
}

// blockIndent returns the indentation of the first non-blank line after the
// block scalar header on line header (1-based).
func blockIndent(lines []string, header int) int {
	for i := header; i < len(lines); i++ {
		trimmed := strings.TrimLeft(lines[i], " ")
		if strings.TrimSpace(trimmed) != "" {
			return len(lines[i]) - len(trimmed)
		}
	}
	return 0
}

func shellDefault(n *yaml.Node) string {
	return scalar(get(get(get(n, "defaults"), "run"), "shell"))
}

// osDefaultShell is pwsh when runs-on names a windows runner and bash
// otherwise, including when runs-on is an expression.
func osDefaultShell(runsOn *yaml.Node) string {
	labels := labelsOf(runsOn)
	for _, l := range labels {
		if strings.Contains(l, "${{") {
			return "bash"
		}
	}
	for _, l := range labels {
		if strings.Contains(strings.ToLower(l), "windows") {
			return "pwsh"
		}
	}
	return "bash"
}

func labelsOf(n *yaml.Node) []string {
	if n == nil {
		return nil
	}
	switch n.Kind {
	case yaml.ScalarNode:
		return []string{n.Value}
	case yaml.SequenceNode:
		var out []string
		for _, c := range n.Content {
			out = append(out, labelsOf(c)...)
		}
		return out
	case yaml.MappingNode:
		return labelsOf(get(n, "labels"))
	}
	return nil
}
