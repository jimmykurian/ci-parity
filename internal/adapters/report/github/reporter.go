package github

import (
	"io"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Reporter implements app.Reporter as GitHub workflow commands.
type Reporter struct {
	W io.Writer
}

// Report writes the findings to r.W. The format has no ratchet block.
func (r Reporter) Report(findings []domain.Finding, _ []domain.Row) error {
	return Write(r.W, findings)
}
