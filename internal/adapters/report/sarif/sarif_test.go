// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

func findings() []domain.Finding {
	return []domain.Finding{
		{Rule: domain.InlineTool, Severity: domain.SeverityError, File: ".github/workflows/ci.yml", Line: 9, Col: 14, Job: "test", Step: "Test", Command: "pnpm test", Message: "error one"},
		{Rule: domain.InlineTool, Severity: domain.SeverityTolerated, File: ".github/workflows/ci.yml", Line: 11, Col: 14, Job: "test", StepIndex: 3, Command: "npm ci", Message: "tolerated one"},
		{Rule: domain.InlineTool, Severity: domain.SeverityAllowed, File: ".github/workflows/ci.yml", Line: 13, Col: 14, Job: "test", Step: "Push", Command: "docker push x", Message: "allowed one", Reason: "images publish only from CI"},
		{Rule: domain.Unsupported, Severity: domain.SeverityWarning, File: ".github/actions/run/action.yml", Line: 5, Col: 9, Action: ".github/actions/run/action.yml", Message: "warning one"},
	}
}

func decode(t *testing.T, fs []domain.Finding, version string) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	if err := Write(&buf, fs, version); err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	return doc
}

func TestRequiredFields(t *testing.T) {
	doc := decode(t, findings(), "0.2.0")
	if doc["version"] != "2.1.0" || doc["$schema"] == nil {
		t.Fatalf("version = %v, $schema = %v", doc["version"], doc["$schema"])
	}
	runs := doc["runs"].([]any)
	if len(runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(runs))
	}
	driver := runs[0].(map[string]any)["tool"].(map[string]any)["driver"].(map[string]any)
	if driver["name"] != "ci-parity" || driver["semanticVersion"] != "0.2.0" || driver["informationUri"] != InformationURI {
		t.Errorf("driver = %v", driver)
	}
	var ids []string
	for _, r := range driver["rules"].([]any) {
		rule := r.(map[string]any)
		for _, key := range []string{"id", "name", "shortDescription", "helpUri", "defaultConfiguration"} {
			if rule[key] == nil {
				t.Errorf("rule %v has no %s", rule["id"], key)
			}
		}
		ids = append(ids, rule["id"].(string))
	}
	if diff := cmp.Diff([]string{"P001", "P002", "P003", "P004", "P005"}, ids); diff != "" {
		t.Errorf("rule ids (-want +got):\n%s", diff)
	}
	for _, r := range runs[0].(map[string]any)["results"].([]any) {
		res := r.(map[string]any)
		loc := res["locations"].([]any)[0].(map[string]any)["physicalLocation"].(map[string]any)
		art := loc["artifactLocation"].(map[string]any)
		if art["uriBaseId"] != "%SRCROOT%" || art["uri"] == "" || loc["region"].(map[string]any)["startLine"] == nil {
			t.Errorf("location = %v", loc)
		}
		if res["partialFingerprints"].(map[string]any)["ciParity/v1"] == "" {
			t.Errorf("result %v has no fingerprint", res["message"])
		}
	}
}

func TestLevelsAndSuppressions(t *testing.T) {
	results := decode(t, findings(), "dev")["runs"].([]any)[0].(map[string]any)["results"].([]any)
	var got []string
	for _, r := range results {
		res := r.(map[string]any)
		entry := res["level"].(string)
		if s, ok := res["suppressions"].([]any); ok {
			sup := s[0].(map[string]any)
			entry += " suppressed:" + sup["kind"].(string) + ":" + sup["justification"].(string)
		}
		got = append(got, entry)
	}
	want := []string{"error", "note", "error suppressed:inSource:images publish only from CI", "warning"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("levels (-want +got):\n%s", diff)
	}
}

func TestSemanticVersionOnlyForReleases(t *testing.T) {
	for version, want := range map[string]any{"0.2.0": "0.2.0", "dev": nil, "0.0.0-SNAPSHOT-abc1234": nil, "v0.2.0": nil} {
		driver := decode(t, nil, version)["runs"].([]any)[0].(map[string]any)["tool"].(map[string]any)["driver"].(map[string]any)
		if driver["semanticVersion"] != want {
			t.Errorf("version %q: semanticVersion = %v, want %v", version, driver["semanticVersion"], want)
		}
	}
}

func TestFingerprintIgnoresPosition(t *testing.T) {
	a := findings()[0]
	b := a
	b.Line, b.Col = 40, 3
	if fingerprint(a) != fingerprint(b) {
		t.Error("moving a finding changed its fingerprint")
	}
	c := a
	c.Command = "pnpm lint"
	if fingerprint(a) == fingerprint(c) {
		t.Error("a different command kept the same fingerprint")
	}
	d := findings()[1]
	e := d
	e.StepIndex = 4
	if fingerprint(d) == fingerprint(e) {
		t.Error("a different unnamed step kept the same fingerprint")
	}
}
