// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package github renders findings as GitHub Actions workflow commands, which
// the runner turns into annotations on the offending lines.
package github

import (
	"fmt"
	"io"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Write emits one workflow command per finding and nothing else. Errors become
// ::error, warnings ::warning and tolerated findings ::notice; allowed
// findings are not emitted.
func Write(w io.Writer, findings []domain.Finding) error {
	var b strings.Builder
	for _, f := range findings {
		level, title := "", f.Rule+" "+f.Name()
		switch f.Severity {
		case domain.SeverityError:
			level = "error"
		case domain.SeverityWarning:
			level = "warning"
		case domain.SeverityTolerated:
			level, title = "notice", title+" (tolerated)"
		default:
			continue
		}
		fmt.Fprintf(&b, "::%s file=%s,line=%d,col=%d,title=%s::%s\n",
			level, property(f.File), f.Line, f.Col, property(title), data(f.Message))
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// data escapes a workflow command message.
func data(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(s)
}

// property escapes a workflow command property value, which additionally
// cannot hold the : and , separators.
func property(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C").Replace(s)
}
