// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Command metrics renders the README's code metrics from scc's JSON on stdin
// (lines of code per file) and the Go files it lists (functions and their
// cyclomatic complexity). `task metrics` prints the block, `task metrics:readme`
// writes it into README.md, and `task metrics:check` fails when it's out of date.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// main renders the block, then prints it, writes it into the README, or
// checks the README already has it.
func main() {
	readme := flag.String("readme", "README.md", "the README holding the METRICS block")
	write := flag.Bool("write", false, "write the block into the README")
	check := flag.Bool("check", false, "fail when the README's block is out of date")
	flag.Parse()
	if err := run(*readme, *write, *check); err != nil {
		fmt.Fprintln(os.Stderr, "metrics:", err)
		os.Exit(1)
	}
}

// run reads scc's output from stdin, renders the block, and prints, writes or
// checks it.
func run(readmePath string, write, check bool) error {
	files, err := ParseSCC(os.Stdin)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("scc counted no files")
	}
	var funcs []Func
	for _, f := range files {
		if !strings.HasSuffix(f.Path, ".go") {
			continue
		}
		src, err := os.ReadFile(f.Path)
		if err != nil {
			return fmt.Errorf("read %s: %w", f.Path, err)
		}
		found, err := ParseFuncs(f.Path, src)
		if err != nil {
			return err
		}
		funcs = append(funcs, found...)
	}
	block := Render(files, funcs)
	if !write && !check {
		fmt.Print(block)
		return nil
	}
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", readmePath, err)
	}
	updated, err := ReplaceBlock(string(data), block)
	if err != nil {
		return err
	}
	if check {
		if updated != string(data) {
			return fmt.Errorf("%s's code metrics are out of date; run `task metrics:readme` and commit the result", readmePath)
		}
		return nil
	}
	return os.WriteFile(readmePath, []byte(updated), 0o644) //nolint:gosec // the README is a public, world-readable file
}
