// Package app holds ci-parity's use cases, check and baseline, and the ports
// they need. Adapters under internal/adapters implement the ports; cmd wires
// them together.
package app

import (
	"github.com/jimmykurian/ci-parity/internal/domain"
	"github.com/jimmykurian/ci-parity/internal/rules"
)

// TaskSource answers questions about the tasks a task runner defines.
type TaskSource = rules.TaskSource

// CommandExtractor finds tool and task-runner calls in a step's script.
type CommandExtractor = rules.CommandExtractor

// ConfigSource loads the ci-parity config.
type ConfigSource interface {
	// Load reads the config at path. An empty path means the default file,
	// whose absence yields the defaults.
	Load(path string) (*domain.Config, error)
}

// WorkflowSource finds and loads workflows and composite actions.
type WorkflowSource interface {
	// Discover returns the repo-relative paths matching any of patterns, sorted.
	Discover(patterns []string) ([]string, error)
	// Load parses the workflow or action at the repo-relative path.
	Load(path string) (*domain.File, error)
}

// TaskLoader opens the task runner's file, which the config names.
type TaskLoader interface {
	// Load reads the runner file at path; an empty path means discovery.
	Load(path string) (TaskSource, error)
}

// BaselineStore reads and writes the baseline file.
type BaselineStore interface {
	// Read loads the baseline at path. A missing file returns an error that
	// wraps fs.ErrNotExist.
	Read(path string) (*domain.Baseline, error)
	// Write saves counts as the baseline at path.
	Write(path string, counts map[string]int) error
}

// Reporter renders the result of a check.
type Reporter interface {
	// Report writes sorted findings and the ratchet rows, which are nil when
	// no baseline applies.
	Report(findings []domain.Finding, rows []domain.Row) error
}

// App runs the use cases against its ports.
type App struct {
	Config    ConfigSource
	Workflows WorkflowSource
	Tasks     TaskLoader
	Extractor CommandExtractor
	Baselines BaselineStore
}
