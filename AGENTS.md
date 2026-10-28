# ci-parity

Checks that GitHub Actions workflows run project commands through the repository's task
runner (Taskfile, make, just or mise) instead of calling tools inline, with a per-workflow
baseline that only shrinks. A Go CLI plus a composite GitHub Action (`action.yml`).

## Map

- `cmd/ci-parity/`: the CLI; `wire.go` builds the adapters and hands them to `app`
- `internal/domain/`: types and pure logic, standard library only
- `internal/rules/`: rules P001 to P005; the ports they use are in `ports.go`
- `internal/app/`: the check, baseline and fix use cases; ports in `ports.go`
- `internal/adapters/`: config, workflow YAML, shell parsing, runner readers, baseline file,
  output formats
- `testdata/fixtures/<case>/`: golden fixtures; `testdata/fix/`: fix fixtures
- `docs/SPEC.md` is the contract, `docs/PLAN.md` the order of work, `docs/DECISIONS.md`
  why things changed

Dependencies point inward: `cmd` → `adapters` → `app` → `rules` → `domain`.

## How to work

- Everything runs through the Taskfile, and `task ci` must pass before a change is done.
  `task test -- -update` regenerates goldens; read the diff, every change must be meant.
- Run `task hooks:install` once. The hooks run gitleaks, the linters and the commit-message
  check, and refuse commits and pushes on `main`: work on a branch and open a pull request.
- Prefer a new fixture to a mock. Fixtures are synthetic (`acme-app`, `widget-service`).
- Work one `docs/PLAN.md` phase at a time, tick its boxes, and stop when it's done. Never
  start anything under Backlog.
- Commit messages and PR titles describe the change, never a plan phase number.
- Keep `README.md` accurate whenever behavior changes.

## Rules no tool checks

- Behavior, rule IDs, messages, config keys and output formats come from `docs/SPEC.md`.
  To change one, or to add a dependency beyond yaml.v3, mvdan.cc/sh/v3, BurntSushi/toml
  and go-cmp, write the proposal in `docs/DECISIONS.md` first, then stop and ask.
- Read runner files statically; never run `task`, `make`, `just` or `mise`. No network
  access at check time, and nothing depends on the environment except `-C`.
- Reported paths are repo-relative with forward slashes; output is sorted by file, line,
  column and rule.
- Wrap errors with context (`fmt.Errorf("parse %s: %w", path, err)`); no panics outside
  `main`.
- Every `run:` step in this repo's workflows calls `task`; setup goes in `uses:` steps,
  pinned to a full commit SHA with a version comment.
- Apache-2.0 lives in `LICENSE`; no per-file headers.
- Fix a gitleaks hit; never allowlist it.

## Read before starting

- `CONTRIBUTING.md`: every task, tests and fixtures, commits and signing, hooks, releases
- `docs/SPEC.md` section 5: how workflows, shells and runner files are read
- `ACCESSIBILITY.md`: rules for output and messages
- `README.md`: user-facing behavior and the JSON and SARIF formats
