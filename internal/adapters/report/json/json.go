// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package json renders findings as the versioned JSON document described in
// docs/reference/output-formats.md.
package json

import (
	encjson "encoding/json"
	"io"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Version is the JSON schema version. It changes only with the schema.
const Version = 1

// document is the whole JSON output.
type document struct {
	Version  int                    `json:"version"`
	Summary  summary                `json:"summary"`
	Findings []finding              `json:"findings"`
	Ratchet  map[string]ratchetItem `json:"ratchet"`
}

// summary counts the findings by severity.
type summary struct {
	Errors    int `json:"errors"`
	Warnings  int `json:"warnings"`
	Tolerated int `json:"tolerated"`
	Allowed   int `json:"allowed"`
}

// finding is one finding in the JSON output. Optional fields are null when
// empty, never omitted.
type finding struct {
	Rule       string  `json:"rule"`
	Name       string  `json:"name"`
	Severity   string  `json:"severity"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Col        int     `json:"col"`
	Job        *string `json:"job"`
	Action     *string `json:"action"`
	Step       *string `json:"step"`
	Command    *string `json:"command"`
	Message    string  `json:"message"`
	Suggestion *string `json:"suggestion"`
	Reason     *string `json:"reason"`
}

// ratchetItem is one workflow's baseline and current inline-tool counts.
type ratchetItem struct {
	Baseline int  `json:"baseline"`
	Current  int  `json:"current"`
	Locked   bool `json:"locked"`
}

// Write renders every finding, including tolerated and allowed ones. rows is
// nil when no baseline is in effect, which renders "ratchet": null.
func Write(w io.Writer, findings []domain.Finding, rows []domain.Row) error {
	doc := document{Version: Version, Findings: make([]finding, 0, len(findings))}
	for _, f := range findings {
		doc.Findings = append(doc.Findings, convert(f))
		switch f.Severity {
		case domain.SeverityError:
			doc.Summary.Errors++
		case domain.SeverityWarning:
			doc.Summary.Warnings++
		case domain.SeverityTolerated:
			doc.Summary.Tolerated++
		case domain.SeverityAllowed:
			doc.Summary.Allowed++
		}
	}
	if rows != nil {
		doc.Ratchet = make(map[string]ratchetItem, len(rows))
		for _, r := range rows {
			doc.Ratchet[r.File] = ratchetItem{Baseline: r.Baseline, Current: r.Current, Locked: r.Locked()}
		}
	}
	enc := encjson.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}

// convert maps a finding to its JSON shape.
func convert(f domain.Finding) finding {
	out := finding{
		Rule: f.Rule, Name: f.Name(), Severity: string(f.Severity), File: f.File,
		Line: f.Line, Col: f.Col, Message: f.Message,
		Action: optional(f.Action), Step: optional(f.Step), Command: optional(f.Command),
		Suggestion: optional(f.Suggestion), Reason: optional(f.Reason),
	}
	if f.Action == "" {
		out.Job = optional(f.Job)
	}
	return out
}

// optional returns nil for an empty string so it encodes as null.
func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
