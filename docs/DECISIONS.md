# Decisions

Changes to the decisions in `AGENTS.md` or the contracts in `docs/SPEC.md`. Each entry
states the problem, the proposed change, and its status (proposed, accepted, rejected).
Proposed entries block work until the owner answers.

## Log

### 2026-10-27: git hooks, and main only through pull requests (accepted)

Problem: the checks ran only in `task ci` and CI, so a forbidden term, a secret or a
malformed commit message was caught after the commit, and work landed on `main` directly.

Change: `.githooks/` holds pre-commit, commit-msg and pre-push hooks, turned on with
`task hooks:install` (the devcontainer runs it). pre-commit refuses commits on `main` and
runs gitleaks, `task scrub` and the linters on the staged files; commit-msg checks
Conventional Commits against the types and scopes in `.gitmessage`; pre-push refuses
pushes to `main` and runs `task test` and `task self-check` when Go code, fixtures, the
module, the Taskfile, the workflows or `.ci-parity*` change. Missing tools are skipped
with a warning, and `--no-verify` bypasses once. Every change now reaches `main` through
a branch and pull request. GitHub branch protection is a separate setting, not changed
here.

### 2026-10-27: linters and formatters beyond golangci-lint and actionlint (accepted)

Problem: SPEC 12 named three dev tools, so YAML, Markdown, shell, spelling, workflow
security and Go vulnerabilities went unchecked, and formatting was plain gofmt.

Change: every tool is pinned in `mise.toml` and has its own `lint:<group>` task;
`task lint` runs them all, and configs live in `.config/linters/`.

- Go: golangci-lint adds copyloopvar, errorlint, gosec, misspell, nolintlint,
  unconvert, unparam and usestdlibvars, and formats with gofumpt and goimports
  (`task fmt` and `fmt:check` now run `golangci-lint fmt`). gosec's G304 is off: reading
  files the user names is the tool's job. govulncheck runs as `lint:vuln`.
- Workflows: zizmor (offline) next to actionlint. Both checkouts now set
  `persist-credentials: false`.
- shellcheck on `*.sh`, yamllint on the repo's YAML (not fixtures), markdownlint-cli2 run
  with `pnx`, typos for spelling, editorconfig-checker for charset, line endings, final
  newlines and trailing whitespace.
- pnpm is pinned at 11.21.0: the 12.x releases fail mise's npm provenance check, and that
  check is not bypassed.

None of these touch `go.mod`, so the three-module rule is unaffected.

### 2026-10-27: add `pnx` to the built-in tools list (accepted)

Problem: pnpm now ships `pnx`, an alias of `pnpm dlx` and `pnpx` that fetches and runs a
package binary. SPEC 4.2's built-in `node` tools list has `npx` and `pnpx` but not
`pnx`, so `run: pnx eslint .` passes unflagged while `run: pnpx eslint .` is P001.

Change: add `pnx` after `pnpx` in the SPEC 4.2 `node` line and in
`domain.BuiltinTools`, and add a `pnx` step to the `inline-tools` fixture so the goldens
cover it. Repos that use `pnx` deliberately can keep it with `tools.remove: [pnx]` or an
`allow:` entry.

### 2026-10-27: pin dev tools in mise.toml (accepted)

Problem: SPEC 12 pins golangci-lint and actionlint only in CI, task floats at `3.x`, and
local installs come from whatever package manager the developer uses. Local lint
results can drift from CI's when a tool upgrades on one side.

Change: a root `mise.toml` pins task, golangci-lint, actionlint and gitleaks. CI's
`jdx/mise-action` step reads it in both jobs, replacing the inline `mise_toml`, the
`go-task/setup-task` steps and the golangci-lint install-only step. Go stays pinned by
`go.mod` alone. gitleaks still runs as its own `gitleaks-action` step; the mise pin is
for local runs. Locally, `mise install` is optional: `task lint` still only checks that
each binary is on `PATH`. No `go.mod` change, so the three-module rule is unaffected.

### 2026-10-27: clean-architecture package layout (accepted)

Problem: AGENTS.md fixed one package per concern, but the core depended on details.
`rules` imported the mvdan-based `shell` and the YAML-based `workflow`, `baseline` mixed
the ratchet math with JSON file I/O, and the check and baseline use cases lived in
`cmd/ci-parity`. A new runner or shell would have meant editing the rules.

Change: dependencies point inward, `cmd -> adapters -> app -> rules -> domain`.

- `internal/domain`: types and pure logic only (findings, the workflow model and
  position mapping, extracted commands, the config value and its globs, the ratchet).
  Standard library only.
- `internal/rules`: P001 to P005. It defines the two interfaces it consumes,
  `TaskSource` (the former `runner.Runner`) and `CommandExtractor`, so it imports only
  `domain`. `CommandExtractor.Supports(shell)` decides which shells are parsed; anything
  else is still P005.
- `internal/app`: the check and baseline use cases and their ports (`ConfigSource`,
  `WorkflowSource`, `TaskLoader`, `BaselineStore`, `Reporter`, plus aliases of the two
  rules interfaces). `TaskLoader` exists because the Taskfile path comes from the config.
- `internal/adapters/{config/yamlconfig,workflow/ghactions,shell/posix,runner/taskfile,
  baseline/jsonfile,report/*}`: the YAML, shell, JSON and output details. Each is
  constructed with the repo root, so `-C` is resolved once in `cmd`.
- `cmd/ci-parity` is the composition root: flags, adapters, exit codes.

Output, exit codes, error messages and the fixtures are unchanged.

### 2026-10-27: wrapper flags that take a value (accepted)

Problem: SPEC 5.3 dropped a wrapper's options but not the values of options that take
one, so `sudo -u deploy docker push` read `deploy` as the command and
`timeout -k 5 30s pnpm test` read `30s`. Both produced no finding.

Change: SPEC 5.3 lists, per built-in wrapper, the flags that consume the next argument.
`env -S` hides the command in a string, so that call is dropped. Wrappers added through
`transparent` in config keep the generic rule; no new config key.

### 2026-10-27: Phase 1 clarifications written into the spec (accepted)

- Unnamed steps read `step 4 in job "web"` (SPEC 6).
- The default `allow` list is empty; the old example was a no-op because `gh` is not a
  tool (SPEC 4.2).
- Text output: singular `1 error` / `1 warning`, column widths, and the `[tolerated]` and
  `[allowed: <reason>]` markers under `--verbose` (SPEC 8).
- Positions are exact for folded and multi-line plain values and after `${{ }}`
  expressions; parse errors are reported at the parser's position (SPEC 5.1, 5.3).
- A YAML line comment after a single-line `run:` value is kept as a shell comment, so
  one-line allow comments work (SPEC 5.1).

### 2026-10-27: Phase 2 spec gaps (accepted)

Approved by the owner and applied to SPEC 5.4 to 5.6. Each item says why the spec as
written would give wrong results.

1. SPEC 5.4, shell invokers. Add: "`-t`, `--taskfile`, `-d`, `--dir` (in any form,
   including `--dir=web` and `-dweb`), `-g` and `--global` select a Taskfile other than
   the root one, so the call is ignored." Why: `task -d web build` names a task in
   `web/Taskfile.yml`; checking it against the root Taskfile reports a false P002.
2. SPEC 5.4, both invokers. Add: "A task name that is not a literal word, or a word in an
   action input that holds `${{ }}`, cannot be checked and is skipped." Why: without it,
   `with: {task: "${{ matrix.task }}"}` splits into `${{`, `matrix.task`, `}}` and reports
   `matrix.task` as unknown.
3. SPEC 5.5, includes. Add: "An included Taskfile with a `default` task also makes the
   bare namespace a valid name." Why: `task docs` runs `docs:default`; without this it is
   a false P002.
4. SPEC 5.5, includes. Replace "(meaning `<dir>/Taskfile.yml`)" with "(the first of the
   root discovery names found there)", and add "An include path holding a template
   (`{{OS}}`) is treated as missing" and "File names are matched exactly against the
   directory listing." Why: task itself searches all names in a directory; templates
   cannot be resolved statically; case-insensitive file systems otherwise report a
   different file name than Linux, which breaks deterministic output.
5. SPEC 5.6. Replace "preferring non-internal tasks" with "never suggesting internal
   tasks". Why: `task` refuses to run an internal task from the command line, so
   suggesting one is wrong advice.
6. SPEC 5.6, P002. Replace the rule with: "A task name that has the unknown name as a
   prefix wins, shortest first. Otherwise suggest the name with the lowest Levenshtein
   distance, when that distance is at most 3 and less than the length of the unknown
   name." Why: the spec does not rank its two kinds of candidate, and plain lowest
   distance suggested `ds` for `test` over `test:unit`.
7. SPEC 5.6, P001. Add: "Ties go to the shorter name, then the alphabetically first.
   Only the canonical task name is suggested, not an alias."

### 2026-10-27: Phase 3 readings of SPEC 4.3 and 8 (accepted, SPEC text unchanged)

Gaps the spec leaves open, filled within its wording:

- Composite actions are scanned files, so they get baseline entries like workflows.
- `ci-parity baseline` with no existing file writes every count without refusing. With a
  file, a workflow missing from it counts as locked at 0, so a new workflow with inline
  tools is a regression. Workflows no longer scanned are dropped on rewrite.
- `--baseline` naming a missing file means strict mode, the same as no file.
- The text ratchet block lists improvements and regressions, as the SPEC 8 example does.
- JSON: `action` is always present (null outside composite actions), `step` is null for
  unnamed steps, `ratchet` lists every scanned file when a baseline applies and is null
  otherwise, and P002's `suggestion` is `task <name>` like P001's.
