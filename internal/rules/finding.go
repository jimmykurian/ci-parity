// Package rules evaluates the ci-parity rules over loaded workflows and
// produces findings shared by every reporter.
package rules

import "sort"

// Severity is how a finding affects the result.
type Severity string

// Severities, per docs/SPEC.md section 3.
const (
	SeverityError     Severity = "error"
	SeverityWarning   Severity = "warning"
	SeverityTolerated Severity = "tolerated"
	SeverityAllowed   Severity = "allowed"
)

// Rule IDs, per docs/SPEC.md section 6.
const (
	InlineTool  = "P001"
	Unsupported = "P005"
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
