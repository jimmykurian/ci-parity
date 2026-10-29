// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package taskfile

import (
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Command is how CI calls a Taskfile task.
func (t *Taskfile) Command() string { return "task" }

// Internal is how P003 describes an internal task.
func (t *Taskfile) Internal() string { return "is marked internal: true" }

// valueFlags are task flags that take the next argument as their value, as
// `task --help` lists them (docs/reference/analysis.md, Invokers).
var valueFlags = map[string]bool{
	"-t": true, "--taskfile": true, "-d": true, "--dir": true,
	"-C": true, "--concurrency": true, "-I": true, "--interval": true, "-o": true, "--output": true,
	"--output-group-begin": true, "--output-group-end": true, "--sort": true, "--temp-dir": true,
	"--timeout": true, "--expiry": true, "--cacert": true, "--cert": true, "--cert-key": true,
	"--remote-cache-dir": true, "--trusted-hosts": true,
}

// noTasks reports whether a task flag makes the call run no task of the root
// Taskfile: it points at another Taskfile, whose tasks ci-parity cannot check,
// or prints a shell completion script.
func noTasks(arg string) bool {
	switch arg {
	case "-t", "--taskfile", "-d", "--dir", "-g", "--global", "--completion":
		return true
	}
	return strings.HasPrefix(arg, "--taskfile=") || strings.HasPrefix(arg, "--dir=") || strings.HasPrefix(arg, "--completion=") ||
		(len(arg) > 2 && (strings.HasPrefix(arg, "-t") || strings.HasPrefix(arg, "-d")) && !strings.HasPrefix(arg, "--"))
}

// Names returns the task names among the arguments of a shell invoker call. It
// returns none when the call selects another Taskfile or prints a completion
// script. Names that are not
// literal, such as ${{ inputs.task }}, cannot be checked and are skipped.
func (t *Taskfile) Names(args []domain.Word) []domain.Word {
	var names []domain.Word
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case !a.Literal:
		case a.Text == "--":
			return names
		case noTasks(a.Text):
			return nil
		case valueFlags[a.Text]:
			i++
		case strings.HasPrefix(a.Text, "-"), strings.Contains(a.Text, "="):
		default:
			names = append(names, a)
		}
	}
	return names
}
