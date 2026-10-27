// Command ci-parity checks that GitHub Actions workflows run project commands
// through the repository's task runner instead of calling tools inline.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// Exit codes per docs/SPEC.md section 9.
const (
	exitOK       = 0
	exitFindings = 1
	exitUsage    = 2
)

// Build metadata, injected with -ldflags "-X main.version=... -X main.commit=... -X main.date=...".
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const usageText = `Usage:
  ci-parity check    [--config PATH] [--baseline PATH] [--format text|github|json]
                     [--strict] [--verbose] [-C DIR]
  ci-parity baseline [--config PATH] [--baseline PATH] [--allow-regression] [-C DIR]
  ci-parity version
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches a subcommand and returns the process exit code. The golden
// test harness calls it in-process.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usageText)
		return exitUsage
	}
	switch args[0] {
	case "check":
		return runCheck(args[1:], stderr)
	case "baseline":
		return runBaseline(args[1:], stderr)
	case "version":
		return runVersion(args[1:], stdout, stderr)
	case "help", "-h", "-help", "--help":
		fmt.Fprint(stdout, usageText)
		return exitOK
	default:
		fmt.Fprintf(stderr, "ci-parity: unknown command %q\n\n%s", args[0], usageText)
		return exitUsage
	}
}

func runCheck(args []string, stderr io.Writer) int {
	fs := newFlagSet("check", stderr)
	fs.String("config", "", "path to the config file")
	fs.String("baseline", "", "path to the baseline file")
	fs.String("format", "text", "output format: text, github or json")
	fs.Bool("strict", false, "ignore the baseline")
	fs.Bool("verbose", false, "include tolerated and allowed findings in text output")
	fs.String("C", "", "change to `DIR` before doing anything")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	return notImplemented("check", stderr)
}

func runBaseline(args []string, stderr io.Writer) int {
	fs := newFlagSet("baseline", stderr)
	fs.String("config", "", "path to the config file")
	fs.String("baseline", "", "path to the baseline file")
	fs.Bool("allow-regression", false, "allow baseline counts to rise")
	fs.String("C", "", "change to `DIR` before doing anything")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	return notImplemented("baseline", stderr)
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("version", stderr)
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	fmt.Fprintf(stdout, "ci-parity %s (%s, %s)\n", version, commit, date)
	return exitOK
}

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
			return exitOK, false
		}
		return exitUsage, false
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(fs.Output(), "%s: unexpected argument %q\n", fs.Name(), fs.Arg(0))
		return exitUsage, false
	}
	return exitOK, true
}

func notImplemented(name string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "ci-parity %s: not implemented\n", name)
	return exitUsage
}
