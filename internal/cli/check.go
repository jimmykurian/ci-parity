// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/jimmykurian/ci-parity/internal/app"
)

// runCheck runs `ci-parity check`: it reports findings in the chosen format and
// returns the exit code.
func runCheck(args []string, stdout, stderr io.Writer) int {
	flags := newFlagSet("check", stderr)
	configPath := flags.String("config", "", "path to the config file")
	baselinePath := flags.String("baseline", "", "path to the baseline file")
	format := flags.String("format", "text", "output format: text, github, json or sarif")
	strict := flags.Bool("strict", false, "ignore the baseline")
	verbose := flags.Bool("verbose", false, "include tolerated and allowed findings in text output")
	dir := flags.String("C", ".", "run as if started in `DIR`")
	if code, ok := parseFlags(flags, args); !ok {
		return code
	}
	if *format != "text" && *format != "github" && *format != "json" && *format != "sarif" {
		fmt.Fprintf(stderr, "ci-parity check: unknown format %q; use text, github, json or sarif\n", *format)
		return ExitUsage
	}
	a, err := newApp(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return ExitUsage
	}
	res, err := a.Check(context.Background(), app.CheckOptions{
		ConfigPath:   *configPath,
		BaselinePath: *baselinePath,
		Strict:       *strict,
		Reporter:     reporter(stdout, *format, *verbose),
	})
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return ExitUsage
	}
	if res.Failed() {
		return ExitFindings
	}
	return ExitOK
}
