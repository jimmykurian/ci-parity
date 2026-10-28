// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"io"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Reporter implements app.Reporter in the text format.
type Reporter struct {
	W       io.Writer
	Verbose bool // include tolerated and allowed findings
}

// Report writes the findings and ratchet rows to r.W.
func (r Reporter) Report(findings []domain.Finding, rows []domain.Row) error {
	return Write(r.W, findings, rows, r.Verbose)
}
