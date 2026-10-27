// Package runner defines what the rules need to know about a task runner, so
// readers for Make, just or mise can be added without touching the rules.
package runner

// Runner answers questions about the tasks a repository's task runner defines.
type Runner interface {
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
