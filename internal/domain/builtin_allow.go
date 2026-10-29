// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

import "strings"

// Reasons for the built-in allow set, shown on the findings it allows.
const (
	ReasonVersion = "built-in: prints a version or environment"
	ReasonInstall = "built-in: installs a tool pinned to a version"
	ReasonAuth    = "built-in: signs in or selects an account"
	ReasonScript  = "built-in: runs an inline script"
)

// BuiltinAllow reports whether a normalized command is in the built-in allow
// set, setup and info commands nobody would move into a task, and why. It is
// checked after the config's own allow entries, unless builtin_allow is false.
func BuiltinAllow(command string) (string, bool) {
	w := strings.Fields(command)
	if len(w) == 0 {
		return "", false
	}
	switch {
	case version(w):
		return ReasonVersion, true
	case pinnedInstall(w):
		return ReasonInstall, true
	case auth(w):
		return ReasonAuth, true
	case inlineScript(w):
		return ReasonScript, true
	}
	return "", false
}

// version matches `<tool> --version`, `<tool> version` and `<tool> -V`, the
// first two with only flags after them, and `go env`.
func version(w []string) bool {
	if len(w) < 2 {
		return false
	}
	if w[0] == "go" && w[1] == "env" {
		return true
	}
	if w[1] == "-V" {
		return len(w) == 2
	}
	if w[1] != "--version" && w[1] != "version" {
		return false
	}
	return allFlags(w[2:])
}

// pinnedInstall matches installs of tools named with an exact version:
// `go install x@v1`, `cargo install x@1` or `cargo install x --version 1`,
// `npm install -g x@1` (also `i` and `--global`), `pnpm add -g x@1`, and
// `pip install x==1` (also pip3 and `python -m pip`). Every package must be
// pinned, and a project install (`-r`, `-e`, a path) never matches.
func pinnedInstall(w []string) bool {
	switch {
	case len(w) >= 3 && w[0] == "go" && w[1] == "install":
		return pinned(w[2:], "@", nil)
	case len(w) >= 3 && w[0] == "cargo" && w[1] == "install":
		return cargoPinned(w[2:])
	case len(w) >= 4 && w[0] == "npm" && (w[1] == "install" || w[1] == "i"),
		len(w) >= 4 && w[0] == "pnpm" && w[1] == "add":
		return global(w[2:]) && pinned(w[2:], "@", nil)
	case len(w) >= 3 && (w[0] == "pip" || w[0] == "pip3") && w[1] == "install":
		return pinned(w[2:], "==", []string{"-r", "--requirement", "-e", "--editable", "-c", "--constraint"})
	case len(w) >= 5 && (w[0] == "python" || w[0] == "python3") && w[1] == "-m" && w[2] == "pip" && w[3] == "install":
		return pinned(w[4:], "==", []string{"-r", "--requirement", "-e", "--editable", "-c", "--constraint"})
	}
	return false
}

// pinned reports whether args name at least one package, every package holds
// sep followed by a version other than latest, and no flag in refuse is given.
// Flags are skipped; a scoped npm package's leading @ is not the separator.
func pinned(args []string, sep string, refuse []string) bool {
	n := 0
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			for _, r := range refuse {
				if a == r || strings.HasPrefix(a, r+"=") {
					return false
				}
			}
			continue
		}
		i := strings.LastIndex(a, sep)
		if i <= 0 || a[i+len(sep):] == "" || a[i+len(sep):] == "latest" {
			return false
		}
		n++
	}
	return n > 0
}

// cargoPinned matches `cargo install` of crates written `name@version`, or of
// one crate with `--version`. Other flags are skipped; `--path` and `--git`
// install from source, so they never match.
func cargoPinned(args []string) bool {
	var crates []string
	versioned := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--path" || a == "--git" || strings.HasPrefix(a, "--path=") || strings.HasPrefix(a, "--git="):
			return false
		case a == "--version":
			versioned, i = i+1 < len(args), i+1
		case strings.HasPrefix(a, "--version="):
			versioned = true
		case strings.HasPrefix(a, "-"):
		default:
			crates = append(crates, a)
		}
	}
	if versioned {
		return len(crates) == 1 && !strings.Contains(crates[0], "@")
	}
	return pinned(crates, "@", nil)
}

// global reports whether an npm or pnpm install is global.
func global(args []string) bool {
	for _, a := range args {
		if a == "-g" || a == "--global" {
			return true
		}
	}
	return false
}

// authCommands are the sign-in and account commands, as their leading words.
var authCommands = [][]string{
	{"docker", "login"},
	{"docker", "logout"},
	{"aws", "configure"},
	{"az", "login"},
	{"az", "account", "set"},
	{"gcloud", "auth"},
	{"gcloud", "config", "set"},
}

// auth matches a command that starts with one of authCommands.
func auth(w []string) bool {
	for _, a := range authCommands {
		if len(w) >= len(a) && equalWords(w[:len(a)], a) {
			return true
		}
	}
	return false
}

// inlineScript matches `python -c`, `python3 -c`, `node -e` and `node --eval`
// with the script after them, and python, python3 or node reading a script
// from standard input: alone, or with `-` as the script.
func inlineScript(w []string) bool {
	switch w[0] {
	case "python", "python3":
		return len(w) == 1 || w[1] == "-c" || (len(w) == 2 && w[1] == "-")
	case "node":
		return len(w) == 1 || w[1] == "-e" || w[1] == "--eval" || (len(w) == 2 && w[1] == "-")
	}
	return false
}

// allFlags reports whether every word starts with -.
func allFlags(w []string) bool {
	for _, a := range w {
		if !strings.HasPrefix(a, "-") {
			return false
		}
	}
	return true
}

// equalWords reports whether a and b hold the same words.
func equalWords(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
