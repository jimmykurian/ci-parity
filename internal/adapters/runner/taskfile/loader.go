package taskfile

import "github.com/jimmykurian/ci-parity/internal/rules"

// Loader implements app.TaskLoader for Taskfiles in the repository at Root.
type Loader struct {
	Root string
}

// Load reads the Taskfile at path, or discovers it when path is empty.
func (l Loader) Load(path string) (rules.TaskSource, error) {
	t, err := Load(l.Root, path)
	if err != nil {
		return nil, err
	}
	return t, nil
}
