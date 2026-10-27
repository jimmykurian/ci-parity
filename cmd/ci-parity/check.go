package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/jimmykurian/ci-parity/internal/baseline"
	"github.com/jimmykurian/ci-parity/internal/report/github"
	jsonreport "github.com/jimmykurian/ci-parity/internal/report/json"
	"github.com/jimmykurian/ci-parity/internal/report/text"
	"github.com/jimmykurian/ci-parity/internal/rules"
)

func runCheck(args []string, stdout, stderr io.Writer) int {
	flags := newFlagSet("check", stderr)
	configPath := flags.String("config", "", "path to the config file")
	baselinePath := flags.String("baseline", "", "path to the baseline file")
	format := flags.String("format", "text", "output format: text, github or json")
	strict := flags.Bool("strict", false, "ignore the baseline")
	verbose := flags.Bool("verbose", false, "include tolerated and allowed findings in text output")
	dir := flags.String("C", ".", "run as if started in `DIR`")
	if code, ok := parseFlags(flags, args); !ok {
		return code
	}
	if *format != "text" && *format != "github" && *format != "json" {
		fmt.Fprintf(stderr, "ci-parity check: unknown format %q; use text, github or json\n", *format)
		return exitUsage
	}
	res, err := analyze(*dir, *configPath, *baselinePath)
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	findings, rows := res.findings, []baseline.Row(nil)
	if !*strict {
		b, err := baseline.Read(*dir, res.baseline)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			// No baseline file: strict mode, every P001 is an error.
		case err != nil:
			fmt.Fprintf(stderr, "ci-parity: %v\n", err)
			return exitUsage
		default:
			findings, rows = baseline.Apply(b, res.paths, findings)
		}
	}
	if err := report(stdout, *format, findings, rows, *verbose); err != nil {
		fmt.Fprintf(stderr, "ci-parity: write report: %v\n", err)
		return exitUsage
	}
	for _, f := range findings {
		if f.Severity == rules.SeverityError {
			return exitFindings
		}
	}
	return exitOK
}

func report(w io.Writer, format string, findings []rules.Finding, rows []baseline.Row, verbose bool) error {
	switch format {
	case "github":
		return github.Write(w, findings)
	case "json":
		return jsonreport.Write(w, findings, rows)
	default:
		return text.Write(w, findings, rows, verbose)
	}
}
