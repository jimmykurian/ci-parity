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
