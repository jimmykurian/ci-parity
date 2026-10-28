// Package domain holds ci-parity's types and pure logic: findings, the
// workflow model, extracted commands, the config value and the baseline
// ratchet. It imports only the standard library.
package domain

import "sort"

// Severity is how a finding affects the result.
type Severity string

// Severities, per docs/reference/README.md.
const (
	SeverityError     Severity = "error"
	SeverityWarning   Severity = "warning"
	SeverityTolerated Severity = "tolerated"
	SeverityAllowed   Severity = "allowed"
)

// Rule IDs, per docs/reference/rules.md.
const (
	InlineTool   = "P001"
	UnknownTask  = "P002"
	InternalTask = "P003"
	Regression   = "P004"
	Unsupported  = "P005"
)

// Names maps rule IDs to their names.
var Names = map[string]string{
	"P001": "inline-tool",
	"P002": "unknown-task",
	"P003": "internal-task",
	"P004": "regression",
	"P005": "unsupported",
}

// Finding is one rule hit at one location.
type Finding struct {
	Rule       string
	Severity   Severity
	File       string // repo-relative, forward slashes
	Line, Col  int
	Job        string // job ID; empty for composite actions
	Action     string // action path for composite actions
	Step       string // step name; empty when unnamed
	StepIndex  int
	Command    string // normalized command for P001
	Message    string
	Suggestion string
	Reason     string // allow reason for allowed findings
	Fix        *Edit  // for P001, the rewrite to its suggestion when fix can make it
}

// Edit replaces Old with New at a 1-based line and column of a file.
type Edit struct {
	File      string // repo-relative, forward slashes
	Line, Col int
	Old, New  string
}

// Name returns the rule name, such as "inline-tool".
func (f Finding) Name() string {
	return Names[f.Rule]
}

// Sort orders findings by file, line, column and rule ID.
func Sort(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Col != b.Col {
			return a.Col < b.Col
		}
		return a.Rule < b.Rule
	})
}
