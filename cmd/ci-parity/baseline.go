// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/jimmykurian/ci-parity/internal/app"
)

func runBaseline(args []string, stdout, stderr io.Writer) int {
	flags := newFlagSet("baseline", stderr)
	configPath := flags.String("config", "", "path to the config file")
	baselinePath := flags.String("baseline", "", "path to the baseline file")
	allowRegression := flags.Bool("allow-regression", false, "allow baseline counts to rise")
	dir := flags.String("C", ".", "run as if started in `DIR`")
	if code, ok := parseFlags(flags, args); !ok {
		return code
	}
	a, err := newApp(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	res, err := a.WriteBaseline(context.Background(), app.BaselineOptions{
		ConfigPath:      *configPath,
		BaselinePath:    *baselinePath,
		AllowRegression: *allowRegression,
	})
	var regs *app.RegressionError
	switch {
	case errors.As(err, &regs):
		fmt.Fprintln(stderr, "ci-parity baseline: refusing to raise the baseline:")
		for _, r := range regs.Rows {
			fmt.Fprintf(stderr, "  %s  %d -> %d\n", r.File, r.Baseline, r.Current)
		}
		fmt.Fprintln(stderr, "fix the new inline-tool findings, or rerun with --allow-regression to accept them")
		return exitFindings
	case err != nil:
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	fmt.Fprintf(stdout, "wrote %s: %s, %s\n", res.Path, plural(res.Workflows, "workflow"), plural(res.Findings, "inline-tool finding"))
	return exitOK
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
