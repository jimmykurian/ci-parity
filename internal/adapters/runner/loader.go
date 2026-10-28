// Package runner picks the reader for the configured task runner.
package runner

import (
	"fmt"

	"github.com/jimmykurian/ci-parity/internal/adapters/runner/justfile"
	"github.com/jimmykurian/ci-parity/internal/adapters/runner/makefile"
	"github.com/jimmykurian/ci-parity/internal/adapters/runner/mise"
	"github.com/jimmykurian/ci-parity/internal/adapters/runner/taskfile"
	"github.com/jimmykurian/ci-parity/internal/domain"
	"github.com/jimmykurian/ci-parity/internal/rules"
)

// Loader implements app.TaskLoader for the repository at Root.
type Loader struct {
	Root string
}

// Load reads the runner's file: the Taskfile at path (or discovered when path
// is empty), or the discovered Makefile, justfile or mise config.
func (l Loader) Load(runner, path string) (rules.TaskSource, error) {
	switch runner {
	case domain.RunnerTaskfile:
		return orNil(taskfile.Load(l.Root, path))
	case domain.RunnerMake:
		return orNil(makefile.Load(l.Root))
	case domain.RunnerJust:
		return orNil(justfile.Load(l.Root))
	case domain.RunnerMise:
		return orNil(mise.Load(l.Root))
	}
	return nil, fmt.Errorf("runner %q has no reader", runner)
}

// orNil keeps a typed nil out of the interface when loading fails.
func orNil[T rules.TaskSource](t T, err error) (rules.TaskSource, error) {
	if err != nil {
		return nil, err
	}
	return t, nil
}
