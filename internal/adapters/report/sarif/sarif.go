// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package sarif renders findings as a SARIF 2.1.0 log, described in
// docs/reference/output-formats.md, for GitHub code scanning and other SARIF readers.
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

// InformationURI is the project's home page, linked from the tool.
const InformationURI = "https://github.com/jimmykurian/ci-parity"

// RulesURI is the docs site's rules section; each rule's help page is
// RulesURI + its lowercase ID + "/", such as rules/p001/.
const RulesURI = "https://ci-parity.is-a.dev/rules/"

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
	{domain.UnknownTask, "CI calls a task that the task runner does not define.", "error"},
	{domain.InternalTask, "CI calls a task the task runner marks internal, private or hidden.", "error"},
	{domain.Regression, "A workflow has more inline-tool findings than its baseline allows.", "error"},
	{domain.Unsupported, "A step that ci-parity cannot analyze.", "warning"},
}

// log is the top-level SARIF log object.
type log struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []run  `json:"runs"`
}

// run is one run of one tool: ci-parity and its results.
type run struct {
	Tool    tool     `json:"tool"`
	Results []result `json:"results"`
}

// tool describes the tool that produced the run.
type tool struct {
	Driver driver `json:"driver"`
}

// driver is the tool's main component: its name, version and rules.
type driver struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	SemanticVersion string `json:"semanticVersion,omitempty"`
	InformationURI  string `json:"informationUri"`
	Rules           []rule `json:"rules"`
}

// rule is a reportingDescriptor: one rule's ID, name, description and default level.
type rule struct {
	ID                   string        `json:"id"`
	Name                 string        `json:"name"`
	ShortDescription     text          `json:"shortDescription"`
	HelpURI              string        `json:"helpUri"`
	DefaultConfiguration configuration `json:"defaultConfiguration"`
}

// configuration is a rule's default configuration; only its level is set.
type configuration struct {
	Level string `json:"level"`
}

// text is a SARIF message: plain text only.
type text struct {
	Text string `json:"text"`
}

// result is one finding: its rule, level, message, location and fingerprint, and a
// suppression when it is allowed.
type result struct {
	RuleID              string            `json:"ruleId"`
	Level               string            `json:"level"`
	Message             text              `json:"message"`
	Locations           []location        `json:"locations"`
	PartialFingerprints map[string]string `json:"partialFingerprints"`
	Suppressions        []suppression     `json:"suppressions,omitempty"`
}

// location wraps where a result was found.
type location struct {
	PhysicalLocation physicalLocation `json:"physicalLocation"`
}

// physicalLocation is a file and a region in it.
type physicalLocation struct {
	ArtifactLocation artifactLocation `json:"artifactLocation"`
	Region           region           `json:"region"`
}

// artifactLocation is a repo-relative path, resolved against the %SRCROOT% base.
type artifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId"`
}

// region is where in the file a result starts. Lines and columns are 1-based.
type region struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}

// suppression marks an allowed finding as suppressed in source, with its reason.
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
			HelpURI: helpURI(r.id), DefaultConfiguration: configuration{r.level},
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

// convert maps a finding to a SARIF result.
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

// helpURI returns the docs page for rule id, which code scanning links from each alert.
func helpURI(id string) string {
	return RulesURI + strings.ToLower(id) + "/"
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
