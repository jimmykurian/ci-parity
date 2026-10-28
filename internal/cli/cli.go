// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package cli is the ci-parity command line: it parses each subcommand's flags,
// builds the adapters over the repository at -C, runs the use case and maps the
// result to an exit code. cmd/ci-parity calls Run; the integration tests call it
// in-process.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

// Exit codes, per docs/reference/cli.md.
const (
	ExitOK       = 0 // no error findings
	ExitFindings = 1 // error findings, or a refused baseline
	ExitUsage    = 2 // a usage, config or file error
)

// Build metadata, injected with -ldflags
// "-X github.com/jimmykurian/ci-parity/internal/cli.version=..." (and .commit, .date).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const usageText = `Usage:
  ci-parity check    [--config PATH] [--baseline PATH] [--format text|github|json|sarif]
                     [--strict] [--verbose] [-C DIR]
  ci-parity baseline [--config PATH] [--baseline PATH] [--allow-regression] [-C DIR]
  ci-parity fix      [--config PATH] [--dry-run] [-C DIR]
  ci-parity version
`

// Run dispatches a subcommand with args (without the program name) and returns
// the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usageText)
		return ExitUsage
	}
	switch args[0] {
	case "check":
		return runCheck(args[1:], stdout, stderr)
	case "baseline":
		return runBaseline(args[1:], stdout, stderr)
	case "fix":
		return runFix(args[1:], stdout, stderr)
	case "version":
		return runVersion(args[1:], stdout, stderr)
	case "help", "-h", "-help", "--help":
		fmt.Fprint(stdout, usageText)
		return ExitOK
	default:
		fmt.Fprintf(stderr, "ci-parity: unknown command %q\n\n%s", args[0], usageText)
		return ExitUsage
	}
}

// runVersion prints the version, commit and build date stamped in at release.
func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("version", stderr)
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	fmt.Fprintf(stdout, "ci-parity %s (%s, %s)\n", version, commit, date)
	return ExitOK
}

// newFlagSet returns a flag set for a subcommand that reports errors to stderr
// instead of exiting.
func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("ci-parity "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

// parseFlags parses args and rejects positional arguments. It reports the exit
// code to return and false when the caller should stop.
func parseFlags(fs *flag.FlagSet, args []string) (int, bool) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK, false
		}
		return ExitUsage, false
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(fs.Output(), "%s: unexpected argument %q\n", fs.Name(), fs.Arg(0))
		return ExitUsage, false
	}
	return ExitOK, true
}
