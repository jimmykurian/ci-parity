// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jimmykurian/ci-parity/internal/adapters/baseline/jsonfile"
	"github.com/jimmykurian/ci-parity/internal/adapters/config/yamlconfig"
	"github.com/jimmykurian/ci-parity/internal/adapters/report/github"
	jsonreport "github.com/jimmykurian/ci-parity/internal/adapters/report/json"
	"github.com/jimmykurian/ci-parity/internal/adapters/report/sarif"
	"github.com/jimmykurian/ci-parity/internal/adapters/report/text"
	"github.com/jimmykurian/ci-parity/internal/adapters/runner"
	"github.com/jimmykurian/ci-parity/internal/adapters/shell/posix"
	"github.com/jimmykurian/ci-parity/internal/adapters/workflow/ghactions"
	"github.com/jimmykurian/ci-parity/internal/app"
)

// newApp builds the use cases over adapters rooted at dir, which stands in
// for -C: paths are resolved against it rather than changing the process
// directory, so in-process tests stay independent.
func newApp(dir string) (*app.App, error) {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("-C %s: not a directory", dir)
	}
	return &app.App{
		Config:    yamlconfig.Source{Root: dir},
		Workflows: ghactions.Source{Root: dir},
		Tasks:     runner.Loader{Root: dir},
		Extractor: posix.Extractor{},
		Baselines: jsonfile.Store{Root: dir},
		Editor:    ghactions.Source{Root: dir},
	}, nil
}

// reporter returns the reporter for a format already validated by the caller.
func reporter(w io.Writer, format string, verbose bool) app.Reporter {
	switch format {
	case "github":
		return github.Reporter{W: w}
	case "json":
		return jsonreport.Reporter{W: w}
	case "sarif":
		return sarif.Reporter{W: w, Version: version}
	default:
		return text.Reporter{W: w, Verbose: verbose}
	}
}
