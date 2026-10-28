// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"io"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Reporter implements app.Reporter as a SARIF 2.1.0 log.
type Reporter struct {
	W       io.Writer
	Version string // the ci-parity build version, for tool.driver
}

// Report writes the findings to r.W. SARIF has no ratchet block.
func (r Reporter) Report(findings []domain.Finding, _ []domain.Row) error {
	return Write(r.W, findings, r.Version)
}
