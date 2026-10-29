// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package rules

import "github.com/jimmykurian/ci-parity/internal/domain"

// TaskSource answers questions about the tasks a repository's task runner
// defines, so readers for other runners can be added without touching the rules.
type TaskSource interface {
	// File returns the runner's root file, repo-relative with forward slashes.
	File() string
	// Tasks returns every name CI can invoke, including aliases, sorted.
	Tasks() []string
	// IsInternal reports whether the task called name is marked internal.
	IsInternal(name string) bool
	// Commands returns the command strings of the task defined under name.
	// Aliases and namespace shortcuts return none, so suggestions name the task.
	Commands(name string) []string
	// Dir returns the directory the task called name runs its commands in,
	// repo-relative with forward slashes and "" for the root, or false when it
	// can't be known statically, such as a templated dir. Suggestions need it
	// to match the step's working directory.
	Dir(name string) (string, bool)
	// Reaches reports whether calling the runner from dir, repo-relative, reads
	// this root file, so a task named there is one of these tasks.
	Reaches(dir string) bool
	// Command returns how CI calls a task, such as "task" or "mise run"; messages
	// and suggestions use it.
	Command() string
	// Internal returns how P003 describes an internal task, in the runner's own
	// terms, such as "is marked internal: true" or "is private".
	Internal() string
	// Names returns the task names among a shell invoker call's arguments, and
	// none when the call targets another file or runs no task.
	Names(args []domain.Word) []domain.Word
}

// CommandExtractor finds tool and task-runner calls in a step's script.
type CommandExtractor interface {
	// Supports reports whether scripts for the step shell sh can be parsed.
	// Steps in any other shell are reported as P005.
	Supports(sh string) bool
	// Extract returns the calls in script. A script that does not parse
	// returns a *domain.ParseError.
	Extract(script string, opts domain.ExtractOptions) (domain.Extraction, error)
}
