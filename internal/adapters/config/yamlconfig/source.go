package yamlconfig

import "github.com/jimmykurian/ci-parity/internal/domain"

// Source implements app.ConfigSource for the repository at Root.
type Source struct {
	Root string
}

// Load reads the config at path. An empty path reads DefaultPath, whose
// absence yields the defaults; a named file must exist.
func (s Source) Load(path string) (*domain.Config, error) {
	if path == "" {
		return Load(s.Root, DefaultPath, false)
	}
	return Load(s.Root, path, true)
}
