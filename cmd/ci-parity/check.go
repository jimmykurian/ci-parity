package main

import (
	"context"
	"fmt"
	"io"

	"github.com/jimmykurian/ci-parity/internal/app"
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
	a, err := newApp(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	res, err := a.Check(context.Background(), app.CheckOptions{
		ConfigPath:   *configPath,
		BaselinePath: *baselinePath,
		Strict:       *strict,
		Reporter:     reporter(stdout, *format, *verbose),
	})
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	if res.Failed() {
		return exitFindings
	}
	return exitOK
}
