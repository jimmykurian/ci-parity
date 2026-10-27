package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/jimmykurian/ci-parity/internal/adapters/baseline/jsonfile"

	"github.com/jimmykurian/ci-parity/internal/domain"
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
	res, err := analyze(*dir, *configPath, *baselinePath)
	if err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	counts := domain.Counts(res.paths, res.findings)
	old, err := jsonfile.Read(*dir, res.baseline)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	if old != nil && !*allowRegression {
		if regs := domain.Regressions(old, counts); len(regs) > 0 {
			fmt.Fprintln(stderr, "ci-parity baseline: refusing to raise the baseline:")
			for _, r := range regs {
				fmt.Fprintf(stderr, "  %s  %d -> %d\n", r.File, r.Baseline, r.Current)
			}
			fmt.Fprintln(stderr, "fix the new inline-tool findings, or rerun with --allow-regression to accept them")
			return exitFindings
		}
	}
	if err := jsonfile.Write(*dir, res.baseline, counts); err != nil {
		fmt.Fprintf(stderr, "ci-parity: %v\n", err)
		return exitUsage
	}
	total := 0
	for _, n := range counts {
		total += n
	}
	fmt.Fprintf(stdout, "wrote %s: %s, %s\n", res.baseline, plural(len(counts), "workflow"), plural(total, "inline-tool finding"))
	return exitOK
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
