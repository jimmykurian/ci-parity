# ci-parity

A Go CLI and composite GitHub Action that checks GitHub Actions workflows run project
commands through the repository's task runner (Taskfile in v0.1) instead of calling tools
inline, and keeps a baseline that can only shrink.

`docs/SPEC.md` says what to build. `docs/PLAN.md` says in what order. This file says how
to work in this repo. Read all three before writing code.

## Commands

Everything runs through the Taskfile. CI calls the same tasks, and the tool checks this
repo's own workflows, so the repo is its own first customer.

- `task build` builds `./bin/ci-parity`
- `task test` runs `go test ./...`; `task test -- -update` regenerates golden files
- `task lint` runs golangci-lint and actionlint
- `task fmt` formats; `task fmt:check` fails on unformatted code
- `task self-check` runs the tool against this repo
- `task scrub` greps the tree for terms in `$SCRUB_TERMS_FILE` (see Clean room)
- `task ci` runs fmt:check, lint, test, build, self-check, scrub

`task ci` must be green before you call any step done.

## Decisions already made (do not relitigate)

- Go, latest stable, `go` directive at least 1.24. Module path `github.com/OWNER/ci-parity`
  where OWNER comes from `git remote get-url origin`. If there is no remote, ask.
- License Apache-2.0. One LICENSE file at the root; no per-file copyright headers.
- Dependencies: `gopkg.in/yaml.v3`, `mvdan.cc/sh/v3`, and `github.com/google/go-cmp` for
  tests. Nothing else without writing the reason in `docs/DECISIONS.md` first.
- CLI uses the standard library `flag` package with subcommand dispatch. No cobra, no viper.
- Taskfile is read statically. Do not shell out to `task` at check time or in tests.
- Baseline format and rule IDs are fixed by `docs/SPEC.md`. Changing them is a spec change.

If a decision turns out to be wrong, write the problem and the proposed change in
`docs/DECISIONS.md`, then stop and ask. Do not quietly work around it.

## Code conventions

- One package per concern under `internal/`: `config`, `workflow`, `shell`, `runner`,
  `rules`, `baseline`, `report`. `cmd/ci-parity` only wires flags to packages.
- Errors are wrapped with context (`fmt.Errorf("parse %s: %w", path, err)`). No panics
  outside `main` for truly impossible states.
- Reported paths are relative to the repo root with forward slashes on every OS.
- Output is deterministic: findings sorted by file, line, column, rule ID.
- No network access at check time. No environment-dependent behavior except `-C DIR`.
- Exported identifiers get doc comments. Keep functions under about 60 lines.

## Testing

- Golden tests drive the project. Each fixture under `testdata/fixtures/<case>/` is a
  complete miniature repo (config, Taskfile, workflows, optional baseline) plus
  `expected/{text,github,json}.golden` and `expected/exitcode`.
- The harness runs the CLI in-process against every fixture in every format. Adding a
  fixture must not require touching the harness.
- Unit tests for the shell extractor and Taskfile reader live next to the code and use
  table tests. Prefer a new fixture over a mock.
- Fixtures are synthetic. Invent names like `acme-app`, `example-api`, `widget-service`.

## Clean room

This is a from-scratch implementation. Do not read, copy, or adapt code, workflows,
fixtures, or identifiers from any other repository on this machine or elsewhere. Do not
mention other companies, products, or internal ticket IDs anywhere in the tree.

A private list of forbidden terms lives outside the repo at the path in
`$SCRUB_TERMS_FILE`. `task scrub` greps for them and must pass before every commit. If the
variable is unset, `task scrub` prints a warning and passes, so CI stays green; set it
locally before you start. Never add the terms themselves to any tracked file.

gitleaks runs in CI. Treat a hit as a build failure, never as something to allowlist.

## Workflow

- Work one phase of `docs/PLAN.md` at a time. Tick its checkboxes as you go.
- When a phase's acceptance criteria pass and `task ci` is green, commit, summarize what
  changed in a few lines, and stop. Do not start the next phase until asked.
- Conventional commits (`feat:`, `fix:`, `test:`, `docs:`, `chore:`). Small commits.
- Every `run:` step in this repo's workflows must invoke `task`. Setup goes in `uses:`
  steps. Anything else needs a `# ci-parity: allow <reason>` comment on the line.
- Pin third-party actions to a full commit SHA with a version comment.
- Keep this file under 150 lines and `README.md` accurate at the end of every phase.

## Do not

- Do not implement anything listed under Phase 5 (backlog) in `docs/PLAN.md`.
- Do not add a config option that the spec does not list. Propose it in `docs/DECISIONS.md`.
- Do not parse PowerShell or cmd steps; report them as P005 per the spec.
- Do not create `CLAUDE.md` or any other agent instruction file. `AGENTS.md` is the only one.
