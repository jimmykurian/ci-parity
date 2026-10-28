// Package sarif renders findings as a SARIF 2.1.0 log, described in docs/SPEC.md
// section 8, for GitHub code scanning and other SARIF readers.
package sarif

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// InformationURI is the project's home page, linked from the tool and its rules.
const InformationURI = "https://github.com/jimmykurian/ci-parity"

const (
	schemaURI = "https://json.schemastore.org/sarif-2.1.0.json"
	srcRoot   = "%SRCROOT%"
	// fingerprintKey names the fingerprint scheme; bump it if the inputs change.
	fingerprintKey = "ciParity/v1"
)

var semver = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// rules is the driver's rule metadata, in rule ID order.
var rules = []struct {
	id, description string
	level           string
}{
	{domain.InlineTool, "A run: step calls a listed tool directly instead of through the task runner.", "error"},
	{domain.UnknownTask, "CI calls a task that the Taskfile does not define.", "error"},
	{domain.InternalTask, "CI calls a task marked internal: true.", "error"},
	{domain.Regression, "A workflow has more inline-tool findings than its baseline allows.", "error"},
	{domain.Unsupported, "A step that ci-parity cannot analyze.", "warning"},
}

type log struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []run  `json:"runs"`
}

type run struct {
	Tool    tool     `json:"tool"`
	Results []result `json:"results"`
}

type tool struct {
	Driver driver `json:"driver"`
}

type driver struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	SemanticVersion string `json:"semanticVersion,omitempty"`
	InformationURI  string `json:"informationUri"`
	Rules           []rule `json:"rules"`
}

type rule struct {
	ID                   string        `json:"id"`
	Name                 string        `json:"name"`
	ShortDescription     text          `json:"shortDescription"`
	HelpURI              string        `json:"helpUri"`
	DefaultConfiguration configuration `json:"defaultConfiguration"`
}

type configuration struct {
	Level string `json:"level"`
}

type text struct {
	Text string `json:"text"`
}

type result struct {
	RuleID              string            `json:"ruleId"`
	Level               string            `json:"level"`
	Message             text              `json:"message"`
	Locations           []location        `json:"locations"`
	PartialFingerprints map[string]string `json:"partialFingerprints"`
	Suppressions        []suppression     `json:"suppressions,omitempty"`
}

type location struct {
	PhysicalLocation physicalLocation `json:"physicalLocation"`
}

type physicalLocation struct {
	ArtifactLocation artifactLocation `json:"artifactLocation"`
	Region           region           `json:"region"`
}

type artifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId"`
}

type region struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}

type suppression struct {
	Kind          string `json:"kind"`
	Justification string `json:"justification"`
}

// Write renders every finding as one SARIF run. version is the ci-parity build
// version; it is also the semanticVersion when it is a plain X.Y.Z.
func Write(w io.Writer, findings []domain.Finding, version string) error {
	d := driver{Name: "ci-parity", Version: version, InformationURI: InformationURI}
	if semver.MatchString(version) {
		d.SemanticVersion = version
	}
	for _, r := range rules {
		d.Rules = append(d.Rules, rule{
			ID: r.id, Name: domain.Names[r.id], ShortDescription: text{r.description},
			HelpURI: InformationURI + "#rules", DefaultConfiguration: configuration{r.level},
		})
	}
	results := make([]result, 0, len(findings))
	for _, f := range findings {
		results = append(results, convert(f))
	}
	doc := log{Schema: schemaURI, Version: "2.1.0", Runs: []run{{Tool: tool{d}, Results: results}}}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}

func convert(f domain.Finding) result {
	r := result{
		RuleID:  f.Rule,
		Level:   level(f.Severity),
		Message: text{f.Message},
		Locations: []location{{PhysicalLocation: physicalLocation{
			ArtifactLocation: artifactLocation{URI: f.File, URIBaseID: srcRoot},
			Region:           region{StartLine: f.Line, StartColumn: f.Col},
		}}},
		PartialFingerprints: map[string]string{fingerprintKey: fingerprint(f)},
	}
	if f.Severity == domain.SeverityAllowed {
		r.Suppressions = []suppression{{Kind: "inSource", Justification: f.Reason}}
	}
	return r
}

// level maps a severity to a SARIF level. An allowed finding keeps its rule's
// level and is marked suppressed instead.
func level(s domain.Severity) string {
	switch s {
	case domain.SeverityWarning:
		return "warning"
	case domain.SeverityTolerated:
		return "note"
	default:
		return "error"
	}
}

// fingerprint identifies a finding by what it is rather than where, so it
// survives lines moving above it.
func fingerprint(f domain.Finding) string {
	scope := f.Job
	if f.Action != "" {
		scope = f.Action
	}
	step := f.Step
	if step == "" && f.StepIndex > 0 {
		step = "#" + strconv.Itoa(f.StepIndex)
	}
	sum := sha256.Sum256([]byte(strings.Join([]string{f.Rule, f.File, scope, step, f.Command}, "\x00")))
	return hex.EncodeToString(sum[:16])
}
