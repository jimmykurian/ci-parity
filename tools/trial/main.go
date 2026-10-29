// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Command trial runs the field trial: it fetches each repository in the trial
// list at its pinned commit, runs `ci-parity check --format json` on it with
// the repository's runner, and prints a Markdown report of the findings.
// `task trial` runs it. Fetching needs the network; ci-parity itself only
// reads the files.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// main runs the trial and prints the report.
func main() {
	list := flag.String("repos", "test/trial/repos.yml", "the trial list")
	bin := flag.String("bin", "bin/ci-parity", "the ci-parity binary to run")
	cache := flag.String("cache", ".cache/trial", "where the repositories are fetched")
	flag.Parse()
	if err := run(*list, *bin, *cache); err != nil {
		fmt.Fprintln(os.Stderr, "trial:", err)
		os.Exit(1)
	}
}

// run fetches and checks every repository in the list, then prints the report.
func run(listPath, bin, cache string) error {
	data, err := os.ReadFile(listPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", listPath, err)
	}
	repos, err := ParseRepos(data)
	if err != nil {
		return err
	}
	// Absolute, since ci-parity runs with -C in each checkout.
	if bin, err = filepath.Abs(bin); err != nil {
		return fmt.Errorf("find %s: %w", bin, err)
	}
	if cache, err = filepath.Abs(cache); err != nil {
		return fmt.Errorf("find %s: %w", cache, err)
	}
	if err := os.MkdirAll(cache, 0o750); err != nil {
		return fmt.Errorf("create %s: %w", cache, err)
	}
	var results []Result
	for _, r := range repos {
		fmt.Fprintf(os.Stderr, "trial: %s\n", r.Repo)
		dir := filepath.Join(cache, r.Dir())
		if err := fetch(r, dir); err != nil {
			return err
		}
		res, err := check(r, bin, dir, filepath.Join(cache, r.Dir()+".ci-parity.yml"))
		if err != nil {
			return err
		}
		results = append(results, res)
	}
	fmt.Print(Render(results))
	return nil
}

// fetch makes dir a checkout of the repository at its pinned commit, fetching
// only that commit, and does nothing when it already is one.
func fetch(r Repo, dir string) error {
	if head, err := git(dir, "rev-parse", "HEAD"); err == nil && head == r.Commit {
		return nil
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); errors.Is(err, os.ErrNotExist) {
		if _, err := git("", "init", "-q", dir); err != nil {
			return err
		}
	}
	if _, err := git(dir, "fetch", "-q", "--depth", "1", "https://github.com/"+r.Repo+".git", r.Commit); err != nil {
		return err
	}
	_, err := git(dir, "checkout", "-q", "--detach", "FETCH_HEAD")
	return err
}

// git runs git in dir (or the current directory when dir is empty) and
// returns its trimmed output.
func git(dir string, args ...string) (string, error) {
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.Command("git", args...) //nolint:gosec // the repo and commit are validated by ParseRepos
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}

// check writes a config naming the repository's runner outside its checkout,
// runs ci-parity on it, and reads the JSON output. Exit codes 0 and 1 are
// results; anything else is an error.
func check(r Repo, bin, dir, config string) (Result, error) {
	if err := os.WriteFile(config, []byte("runner: "+r.Runner+"\n"), 0o600); err != nil {
		return Result{}, fmt.Errorf("write %s: %w", config, err)
	}
	cmd := exec.Command(bin, "check", "-C", dir, "--config", config, "--format", "json") //nolint:gosec // the binary is this repository's build
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	var exit *exec.ExitError
	if err != nil && (!errors.As(err, &exit) || exit.ExitCode() != 1) {
		return Result{}, fmt.Errorf("check %s: %w: %s", r.Repo, err, strings.TrimSpace(stderr.String()))
	}
	return ParseResult(r, out)
}
