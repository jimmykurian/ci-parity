package ghactions

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Discover returns the repo-relative, forward-slash paths of files under root
// that match any of patterns, sorted and without duplicates.
func Discover(root string, patterns []string) ([]string, error) {
	seen := map[string]bool{}
	for _, pattern := range patterns {
		start := filepath.Join(root, filepath.FromSlash(domain.StaticPrefix(pattern)))
		err := filepath.WalkDir(start, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return nil
				}
				return err
			}
			if d.IsDir() {
				if d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if domain.MatchPath(pattern, rel) {
				seen[rel] = true
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("discover %s: %w", pattern, err)
		}
	}
	paths := make([]string, 0, len(seen))
	for p := range seen {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths, nil
}
