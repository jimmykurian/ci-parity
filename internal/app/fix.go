package app

import (
	"context"
	"errors"
	"io/fs"
	"sort"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// FixOptions configures Fix.
type FixOptions struct {
	ConfigPath string // empty means the default config file, which is optional
	DryRun     bool   // report the edits without writing them
}

// FixResult is what Fix changed, or would change on a dry run.
type FixResult struct {
	Edits    []domain.Edit // sorted by file, line and column
	Files    int           // files changed
	Manual   int           // inline-tool findings fix cannot rewrite
	Baseline bool          // a baseline file exists, so its counts may now drop
}

// Fix rewrites each inline-tool step whose whole run: value is a command a
// task already runs to call the task instead. Allowed findings are left
// alone; the rest that cannot be rewritten safely are counted in Manual.
func (a *App) Fix(ctx context.Context, opts FixOptions) (FixResult, error) {
	an, err := a.analyze(ctx, opts.ConfigPath, "")
	if err != nil {
		return FixResult{}, err
	}
	var res FixResult
	byFile := map[string][]domain.Edit{}
	for _, f := range an.findings {
		if f.Rule != domain.InlineTool || f.Severity == domain.SeverityAllowed {
			continue
		}
		if f.Fix == nil {
			res.Manual++
			continue
		}
		byFile[f.File] = append(byFile[f.File], *f.Fix)
	}
	files := make([]string, 0, len(byFile))
	for file := range byFile {
		files = append(files, file)
	}
	sort.Strings(files)
	for _, file := range files {
		edits := byFile[file]
		if !opts.DryRun {
			planned := len(edits)
			if edits, err = a.Editor.Apply(file, edits); err != nil {
				return FixResult{}, err
			}
			// An edit whose text didn't match the file, such as an escaped quoted
			// value, still needs a hand.
			res.Manual += planned - len(edits)
		}
		if len(edits) > 0 {
			res.Files++
			res.Edits = append(res.Edits, edits...)
		}
	}
	if _, err := a.Baselines.Read(an.baseline); err == nil {
		res.Baseline = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		return FixResult{}, err
	}
	return res, nil
}
