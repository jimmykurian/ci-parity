// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"io"

	"github.com/jimmykurian/ci-parity/internal/app"
)

// runFix runs `ci-parity fix`: it rewrites inline-tool steps to call their task and
// returns the exit code.
func runFix(args []string, stdout, stderr io.Writer) int {
	flags := newFlagSet("fix", stderr)
	configPath := flags.String("config", "", "path to the config file")
	dryRun := flags.Bool("dry-run", false, "show the changes without writing them")
	dir := flags.String("C", ".", "run as if started in `DIR`")
	if code, ok := parseFlags(flags, args); !ok {
		return code
	}
	a, err := newApp(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	res, err := a.Fix(context.Background(), app.FixOptions{ConfigPath: *configPath, DryRun: *dryRun})
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	for _, e := range res.Edits {
		fmt.Fprintf(stdout, "%s:%d:%d '%s' -> '%s'\n", e.File, e.Line, e.Col, e.Old, e.New)
	}
	verb := "fixed"
	if *dryRun {
		verb = "would fix"
	}
	fmt.Fprintf(stdout, "%s %s in %s; %s a manual change\n", verb, plural(len(res.Edits), "step"),
		plural(res.Files, "file"), needs(res.Manual))
	if res.Baseline && len(res.Edits) > 0 && !*dryRun {
		fmt.Fprintln(stdout, "run `ci-parity baseline` to lock in the lower counts")
	}
	return exitOK
}

// needs phrases how many inline-tool findings still need a hand edit.
func needs(n int) string {
	if n == 1 {
		return "1 inline-tool finding needs"
	}
	return fmt.Sprintf("%d inline-tool findings need", n)
}
