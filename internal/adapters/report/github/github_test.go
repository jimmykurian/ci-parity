package github

import (
	"bytes"
	"testing"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

func TestWriteEscapes(t *testing.T) {
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
	if err := Write(&buf, findings); err != nil {
		t.Fatal(err)
	}
	want := "::error file=.github/workflows/a%2Cb%3Ac.yml,line=3,col=9,title=P001 inline-tool::runs 'echo 100%25' directly%0Asecond line, with: colon\n" +
		"::notice file=ci.yml,line=4,col=1,title=P001 inline-tool (tolerated)::m\n" +
		"::warning file=ci.yml,line=6,col=2,title=P005 unsupported::w\n"
	if buf.String() != want {
		t.Errorf("Write =\n%s\nwant\n%s", buf.String(), want)
	}
}
