// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// CheckOptions configures Check.
type CheckOptions struct {
	ConfigPath   string // empty means the default config file, which is optional
	BaselinePath string // empty means the configured baseline
	Strict       bool   // ignore the baseline
	Reporter     Reporter
}

// CheckResult is what a check found.
type CheckResult struct {
	Findings []domain.Finding
	Rows     []domain.Row // nil when no baseline applies
}

// Failed reports whether any finding is an error.
func (r CheckResult) Failed() bool {
	for _, f := range r.Findings {
		if f.Severity == domain.SeverityError {
			return true
		}
	}
	return false
}

// Check runs the rules, applies the baseline unless opts.Strict is set or no
// baseline file exists, and reports the result.
func (a *App) Check(ctx context.Context, opts CheckOptions) (CheckResult, error) {
	an, err := a.analyze(ctx, opts.ConfigPath, opts.BaselinePath)
	if err != nil {
		return CheckResult{}, err
	}
	res := CheckResult{Findings: an.findings}
	if !opts.Strict {
		b, err := a.Baselines.Read(an.baseline)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			// No baseline file: strict mode, every P001 is an error.
		case err != nil:
			return CheckResult{}, err
		default:
			res.Findings, res.Rows = domain.Apply(b, an.paths, an.findings)
		}
	}
	if err := opts.Reporter.Report(res.Findings, res.Rows); err != nil {
		return CheckResult{}, fmt.Errorf("write report: %w", err)
	}
	return res, nil
}
