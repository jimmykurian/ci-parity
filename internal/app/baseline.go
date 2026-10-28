// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
	"io/fs"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// BaselineOptions configures WriteBaseline.
type BaselineOptions struct {
	ConfigPath      string // empty means the default config file, which is optional
	BaselinePath    string // empty means the configured baseline
	AllowRegression bool   // allow counts to rise
}

// BaselineResult describes the baseline that was written.
type BaselineResult struct {
	Path      string
	Workflows int // scanned files recorded
	Findings  int // inline-tool findings recorded across them
}

// RegressionError is returned when the new counts would raise the baseline
// and regressions are not allowed. Nothing is written.
type RegressionError struct {
	Rows []domain.Row
}

func (e *RegressionError) Error() string {
	return "refusing to raise the baseline"
}

// WriteBaseline records the current inline-tool counts as the baseline. With
// an existing baseline it refuses to raise any count unless
// opts.AllowRegression is set.
func (a *App) WriteBaseline(ctx context.Context, opts BaselineOptions) (BaselineResult, error) {
	an, err := a.analyze(ctx, opts.ConfigPath, opts.BaselinePath)
	if err != nil {
		return BaselineResult{}, err
	}
	counts := domain.Counts(an.paths, an.findings)
	old, err := a.Baselines.Read(an.baseline)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return BaselineResult{}, err
	}
	if old != nil && !opts.AllowRegression {
		if regs := domain.Regressions(old, counts); len(regs) > 0 {
			return BaselineResult{}, &RegressionError{Rows: regs}
		}
	}
	if err := a.Baselines.Write(an.baseline, counts); err != nil {
		return BaselineResult{}, err
	}
	total := 0
	for _, n := range counts {
		total += n
	}
	return BaselineResult{Path: an.baseline, Workflows: len(counts), Findings: total}, nil
}
