package domain

import "strings"

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
	With  map[string]Input
}

// Input is one scalar with: value of a uses: step and its position.
type Input struct {
	Value     string
	Line, Col int
}

// Script is a step's run: value and where it sits in the file.
type Script struct {
	Text     string
	Line     int       // source line of the script's first line
	Col      int       // source column of the script's first character
	Indent   int       // indentation of block scalar content
	Block    bool      // literal or folded block scalar
	Segments []Segment // source positions of folded text; nil when not folded
}

// Segment ties a byte offset in a folded script to the source position its
// text came from.
type Segment struct {
	Off, Line, Col int
}

// Position maps a 1-based line and column inside the script to a position in
// the file.
func (s *Script) Position(line, col int) (int, int) {
	switch {
	case s.Segments != nil:
		return s.foldedPosition(line, col)
	case s.Block:
		return s.Line + line - 1, s.Indent + col
	case line == 1:
		return s.Line, s.Col + col - 1
	default:
		return s.Line + line - 1, col
	}
}

// foldedPosition maps a script line and column through the folding segments.
func (s *Script) foldedPosition(line, col int) (int, int) {
	off := 0
	for i := 1; i < line; i++ {
		next := strings.IndexByte(s.Text[off:], '\n')
		if next < 0 {
			break
		}
		off += next + 1
	}
	off += col - 1
	seg := s.Segments[0]
	for _, sg := range s.Segments {
		if sg.Off > off {
			break
		}
		seg = sg
	}
	return seg.Line, seg.Col + off - seg.Off
}
