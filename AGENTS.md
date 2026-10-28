<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# ci-parity

Checks that GitHub Actions workflows run project commands through the repository's task
runner (Taskfile, make, just or mise) instead of calling tools inline, with a per-workflow
baseline that only shrinks. A Go CLI plus a composite GitHub Action (`action.yml`).

## Map

- `cmd/ci-parity/`: `main`, which calls `internal/cli`
- `internal/cli/`: flags, exit codes and `wire.go`, which builds the adapters and hands
  them to `app`
- `internal/domain/`: types and pure logic, standard library only
- `internal/rules/`: rules P001 to P005; the ports they use are in `ports.go`
- `internal/app/`: the check, baseline and fix use cases; ports in `ports.go`
- `internal/adapters/`: config, workflow YAML, shell parsing, runner readers, baseline file,
  output formats
- `testdata/fixtures/<case>/`: golden fixtures; `testdata/fix/`: fix fixtures
- Unit tests sit next to the code; `test/integration/` runs the CLI in-process on the
  fixtures; `test/e2e/` builds the binary and runs it as a process
- `docs/reference/` is the contract; `docs/architecture/` shows how it fits together and
  why; `docs/README.md` indexes the rest

Dependencies point inward: `cmd` → `cli` → `adapters` → `app` → `rules` → `domain`.

## How to work

- Everything runs through the Taskfile, and `task ci` must pass before a change is done.
  `task test -- -update` regenerates goldens; read the diff, every change must be meant.
- Run `task hooks:install` once. The hooks run gitleaks, the linters and the commit-message
  check, and refuse commits and pushes on `main`: work on a branch and open a pull request.
- Prefer a new fixture to a mock. Fixtures are synthetic (`acme-app`, `widget-service`).
- Commit messages and PR titles describe the change.
- Keep `README.md` and `docs/` accurate whenever behavior changes; follow
  `docs/STYLE_GUIDE.md`.

## Rules no tool checks

- Behavior, rule IDs, messages, config keys and output formats are documented in
  `docs/reference/`. To change one, or to add a dependency beyond yaml.v3, mvdan.cc/sh/v3,
  BurntSushi/toml and go-cmp, propose it and stop to ask first; then update the reference
  in the same pull request.
- Read runner files statically; never run `task`, `make`, `just` or `mise`. No network
  access at check time, and nothing depends on the environment except `-C`.
- Reported paths are repo-relative with forward slashes; output is sorted by file, line,
  column and rule.
- Wrap errors with context (`fmt.Errorf("parse %s: %w", path, err)`); no panics outside
  `main`.
- Every `run:` step in this repo's workflows calls `task`; setup goes in `uses:` steps,
  pinned to a full commit SHA with a version comment.
- Every file starts with the Apache-2.0 header (`task lint:headers`; exceptions in
  `.config/linters/headers-exclude.txt`). Every Go type and function has a doc comment.
- Fix a gitleaks hit; never allowlist it.

## Read before starting

- `CONTRIBUTING.md`: every task, tests and fixtures, commits and signing, hooks, releases
- `docs/reference/analysis.md` and `runners.md`: how workflows, shells and runner files
  are read
- `ACCESSIBILITY.md`: rules for output and messages
- `docs/reference/output-formats.md`: the exact text, `github`, JSON and SARIF formats
