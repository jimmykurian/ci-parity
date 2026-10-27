package main

import (
	"fmt"
	"os"

	"github.com/jimmykurian/ci-parity/internal/config"
	"github.com/jimmykurian/ci-parity/internal/rules"
	"github.com/jimmykurian/ci-parity/internal/runner/taskfile"
	"github.com/jimmykurian/ci-parity/internal/workflow"
)

// analysis is the result of running the rules over a repository.
type analysis struct {
	findings []rules.Finding
	paths    []string // every scanned workflow and action, repo-relative
	baseline string   // baseline path, relative to the repo root
}

// analyze loads the config, the Taskfile and every matching workflow under
// root and runs the rules. root stands in for -C: paths are resolved against
// it rather than changing the process directory, so in-process tests stay
// independent. A non-empty baselinePath overrides the configured one.
func analyze(root, configPath, baselinePath string) (*analysis, error) {
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("-C %s: not a directory", root)
	}
	required := configPath != ""
	if !required {
		configPath = config.DefaultPath
	}
	cfg, err := config.Load(root, configPath, required)
	if err != nil {
		return nil, err
	}
	paths, err := workflow.Discover(root, cfg.Workflows)
	if err != nil {
		return nil, err
	}
	files := make([]*workflow.File, 0, len(paths))
	for _, p := range paths {
		f, err := workflow.Load(root, p)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	tf, err := taskfile.Load(root, cfg.Taskfile)
	if err != nil {
		return nil, err
	}
	if baselinePath == "" {
		baselinePath = cfg.Baseline
	}
	return &analysis{findings: rules.Check(cfg, tf, files), paths: paths, baseline: baselinePath}, nil
}
