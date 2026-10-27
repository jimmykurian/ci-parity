package json

import (
	"io"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Reporter implements app.Reporter as the versioned JSON document.
type Reporter struct {
	W io.Writer
}

// Report writes the findings and ratchet rows to r.W.
func (r Reporter) Report(findings []domain.Finding, rows []domain.Row) error {
	return Write(r.W, findings, rows)
}
