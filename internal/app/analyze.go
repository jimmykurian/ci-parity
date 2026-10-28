package app

import (
	"context"

	"github.com/jimmykurian/ci-parity/internal/domain"
	"github.com/jimmykurian/ci-parity/internal/rules"
)

// analysis is the result of running the rules over a repository.
type analysis struct {
	findings []domain.Finding
	paths    []string // every scanned workflow and action, repo-relative
	baseline string   // baseline path, relative to the repo root
}

// analyze loads the config, the task runner and every matching workflow and
// runs the rules. A non-empty baselinePath overrides the configured one.
func (a *App) analyze(ctx context.Context, configPath, baselinePath string) (*analysis, error) {
	cfg, err := a.Config.Load(configPath)
	if err != nil {
		return nil, err
	}
	paths, err := a.Workflows.Discover(cfg.Workflows)
	if err != nil {
		return nil, err
	}
	files := make([]*domain.File, 0, len(paths))
	for _, p := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f, err := a.Workflows.Load(p)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	tasks, err := a.Tasks.Load(cfg.Runner, cfg.Taskfile)
	if err != nil {
		return nil, err
	}
	if baselinePath == "" {
		baselinePath = cfg.Baseline
	}
	return &analysis{findings: rules.Check(cfg, tasks, a.Extractor, files), paths: paths, baseline: baselinePath}, nil
}
