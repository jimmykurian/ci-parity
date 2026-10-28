// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"bytes"
	"testing"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

func TestWriteEscapes(t *testing.T) {
	// Arrange
	findings := []domain.Finding{
		{
			Rule: "P001", Severity: domain.SeverityError, File: ".github/workflows/a,b:c.yml", Line: 3, Col: 9,
			Message: "runs 'echo 100%' directly\nsecond line, with: colon",
		},
		{Rule: "P001", Severity: domain.SeverityTolerated, File: "ci.yml", Line: 4, Col: 1, Message: "m"},
		{Rule: "P001", Severity: domain.SeverityAllowed, File: "ci.yml", Line: 5, Col: 1, Message: "hidden"},
		{Rule: "P005", Severity: domain.SeverityWarning, File: "ci.yml", Line: 6, Col: 2, Message: "w"},
	}
	var buf bytes.Buffer
	want := "::error file=.github/workflows/a%2Cb%3Ac.yml,line=3,col=9,title=P001 inline-tool::runs 'echo 100%25' directly%0Asecond line, with: colon\n" +
		"::notice file=ci.yml,line=4,col=1,title=P001 inline-tool (tolerated)::m\n" +
		"::warning file=ci.yml,line=6,col=2,title=P005 unsupported::w\n"

	// Act
	err := Write(&buf, findings)
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if buf.String() != want {
		t.Errorf("Write =\n%s\nwant\n%s", buf.String(), want)
	}
}
