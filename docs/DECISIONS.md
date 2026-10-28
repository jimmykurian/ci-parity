# Decisions

Changes to the decisions in `AGENTS.md` or the contracts in `docs/SPEC.md`. Each entry
states the problem, the proposed change, and its status (proposed, accepted, rejected).
Proposed entries block work until the owner answers.

## Log

### 2026-10-28: fix mode (accepted)

Problem: P001 often knows the task that replaces an inline step, but every change was
by hand.

Change: `ci-parity fix [--config PATH] [--dry-run] [-C DIR]` rewrites a P001 finding to
its suggestion only when that cannot change what the step does: the step's `run:` is an
unquoted single-line value, the whole value is exactly the flagged command (no wrapper,
variable assignment, operator, quoting or `${{ }}`), a suggestion exists, and the
finding is not allowed. The value is replaced in place, so a trailing YAML comment and
every other byte of the file stay as they were; each edit is checked against the file
before it is written and skipped if the text no longer matches. Every other inline-tool
finding that is not allowed is counted as needing a manual change. Output is one line
per edit, `file:line:col 'old' -> 'new'`, then a summary; with a baseline file, a hint to
run `ci-parity baseline`. `--dry-run` prints the same without writing. Exit codes: 0, or 2
for a usage or input error. The action does not run fix. Accepted on the owner's
instruction to complete the planned phases.

### 2026-10-28: P003 wording per runner (accepted)

Problem: P003 said "is marked internal: true" for every runner, but that is Taskfile
syntax; a just recipe is private (`[private]` or a leading `_`) and a mise task is
hidden (`hide = true`).

Change: P003 reads `'<name>' <marker> and should not be invoked from CI`, where the
marker is the runner's own: `is marked internal: true` for taskfile (unchanged),
`is private` for just and `is hidden` for mise. make has no internal targets.

### 2026-10-28: Make, just and mise readers (accepted)

Problem: only Taskfiles are read, and `make`, `just` and `mise` are always flagged as
bypass tools, so a repository built around one of them cannot use ci-parity.

Change. `runner` accepts `taskfile`, `make`, `just` or `mise`. Each runner has a reader
behind the existing `TaskSource` port and its own call syntax; the rules stay the same.

- Discovery, no new config key: the reader finds its root file in the repository root,
  matching names exactly. `taskfile` keeps its key; setting it with another runner is a
  config error. make: `GNUmakefile`, `makefile`, `Makefile`. just: `justfile`,
  `Justfile`, `.justfile`. mise: `mise.toml`, `.mise.toml`, `mise/config.toml`,
  `.mise/config.toml`, `.config/mise.toml`, `.config/mise/config.toml`, merged in mise's
  order.
- Defaults that follow the runner: `invokers.shell` defaults to `[make, gmake]`,
  `[just]` or `["mise run", "mise r", "mise tasks run"]` (taskfile keeps its three), and
  `invokers.actions` defaults to none outside taskfile. The configured runner leaves the
  tools list; `task` and `go-task` join it as bypasses when the runner is not taskfile.
- Wording: every place that says `task <name>` says the runner's command instead
  (`make lint`, `just lint`, `mise run lint`): the P001 suggestion, the P002 message and
  its suggestion, and the JSON and SARIF `suggestion`. Taskfile output is unchanged.
- make: targets are lines `name...:` outside recipes, skipping special targets
  (`.PHONY`), pattern rules (`%`), variable assignments and names holding `$(...)`;
  `include`/`-include` of literal paths are followed (depth 10, cycles ignored). Recipe
  lines, without leading `@`, `-` or `+`, are the commands for suggestions. No target
  is internal. A call's value flags are `-f --file --makefile -C --directory -I
  --include-dir -j --jobs -l --load-average -o --old-file -W --what-if`; `-f`, `-C` and
  their long forms select another makefile, so the call is ignored; `NAME=value` is a
  variable; `make` alone runs the default goal and is ignored.
- just: recipes, their parameters, `alias a := b`, `import` of literal paths and `mod`
  (recipes named `mod::recipe`, from `name.just`, `name/mod.just`, `name/justfile` or a
  given path). A recipe is internal when it is `[private]` or starts with `_`. Body
  lines without leading `@` or `-` are its commands. In a call, the words after a
  recipe are its arguments, as many as its parameters take (all the rest for a
  variadic one); `just mod recipe` and `just mod::recipe` both name a module recipe.
  `-f`, `--justfile`, `-d` and `--working-directory` select another justfile; `--set`
  takes two values.
- mise: `[tasks.name]` tables and `name = "cmd"` entries, with `run` (string or list) as
  the commands, `alias` as aliases and `hide = true` as internal; file tasks under
  `mise-tasks/`, `.mise-tasks/`, `mise/tasks/`, `.mise/tasks/` and `.config/mise/tasks/`
  (names from the path, `db/migrate` as `db:migrate`, no commands); and
  `task_config.includes`. A call names one task, then its arguments; `:::` starts
  another task. `-C` and `--cd` select another directory.
- Dependency: `github.com/BurntSushi/toml` (MIT, no dependencies of its own) reads mise's
  TOML; Makefiles and justfiles are read by small parsers here. This is the fourth
  module in `go.mod`.
- Tests: table tests for each reader and call parser, and fixtures per runner (P001 to
  P003, suggestions, includes or modules, another-file flags, the bypass list).
- Plan: one phase, built as three pull requests (make, then just, then mise), released
  together as v0.3.0; fix mode becomes v0.4.0.

### 2026-10-28: Phase 6, unwrapping package runners for suggestions (accepted)

Problem: a P001 suggestion needs the inline command to equal a Taskfile command after
whitespace and the `pnpm run`/`yarn run` rule. So `run: npx eslint .` gets no
"already available as 'task lint'" when the task runs `eslint .`, and a task that runs
`uv run pytest` is never suggested for an inline `pytest`. Package runners are how most
repositories call their tools, so most suggestions are missed.

Change (SPEC 5.6, P001): before comparing, both commands are also unwrapped once. A
command that starts with a package runner below drops the runner words and the options
after them, and what remains is compared, still after the `run` rule. An inline command
and a task command match when they are equal as written or unwrapped, on either side.

| Runner words | Options that take a separate value | Not unwrapped when |
| --- | --- | --- |
| `npx`, `pnpx`, `pnx`, `bunx` | `-p --package` | `-c` or `--call` is given |
| `pnpm exec`, `pnpm dlx` | none | |
| `uv run`, `uvx` | `--with --with-editable --with-requirements --python -p --project --directory --extra --group --package --env-file --from --index --default-index` | |
| `dotnet tool run` | none | |

- Options are arguments that start with `-` before the first other argument; `--` ends
  them; `--flag=value` carries its own value. Nothing left after them means the command
  is not unwrapped.
- The backlog named `npx`, `pnpm exec`, `uv run` and `dotnet tool run`. `pnpx`, `pnx`,
  `bunx`, `pnpm dlx` and `uvx` are added because they are the same runners under other
  names.
- Only suggestions change. Detection, the reported `command`, messages, allow matching
  and baselines are untouched; the `suggestion` field and the message suffix appear
  where they were missing.
- Tests: table tests for the unwrapping, and a new `unwrap-suggestions` fixture covering
  each runner in both directions plus `npx -c`, which keeps no suggestion.
- Acceptance: `task ci` green, existing goldens unchanged except where a suggestion was
  missing, and the README's Rules section notes the equivalence. Release v0.3.0.

### 2026-10-27: promote four backlog items into Phases 5 to 8 (accepted)

Problem: `AGENTS.md` forbids the Phase 5 backlog and `docs/PLAN.md` marks it "do not
start", but the owner asked to build SARIF output, unwrapping for suggestions, fix mode,
and Make, just and mise readers.

Change:

- `docs/PLAN.md`: the backlog becomes a "Backlog" section holding the two items not
  chosen (`task --list-all --json` as a source, a companion run-task action), and four
  phases replace it, one item each, each released on its own (a `feat` makes a minor
  release: v0.2.0, v0.3.0, ...): Phase 5 SARIF output, Phase 6 unwrapping, Phase 7 Make,
  just and mise readers, Phase 8 fix mode.
- `AGENTS.md` and `docs/SPEC.md` 2: the "do not implement" rule and the out-of-scope list
  name the Backlog section instead of Phase 5; the README FAQ, CONTRIBUTING and the
  feature-request form follow.
- Each phase's spec text is proposed here when that phase starts, and approved before
  its code. Phase 5's:

Phase 5, SARIF output (a new SPEC 8 subsection):

- `check --format sarif` prints one SARIF 2.1.0 log with one run, deterministic and
  sorted like the other formats. Exit codes are unchanged.
- `runs[0].tool.driver`: `name` `ci-parity`, `version` and `semanticVersion` from the
  build, `informationUri` the repository URL, and a `rules` entry per rule P001 to P005
  with `id`, `name`, `shortDescription`, `helpUri` (the README's Rules section) and
  `defaultConfiguration.level`.
- One result per finding: `ruleId`, `level` `error` for errors, `warning` for warnings
  and `note` for tolerated findings; an allowed finding is a result with
  `suppressions: [{"kind": "inSource", "justification": <reason>}]`, so code scanning
  shows it as dismissed rather than dropping it. `message.text` is the same message as
  the other formats.
- Location: `artifactLocation.uri` is the repository-relative path with `uriBaseId`
  `%SRCROOT%`; `region.startLine` and `startColumn` are the finding's position.
- `partialFingerprints["ciParity/v1"]` is a hash of rule, file, job (or action), step
  and command, so a finding keeps its identity when lines above it move.
- The action: with `format: sarif` it writes the log to `$RUNNER_TEMP/ci-parity.sarif`
  and sets a new `sarif-file` output; the README shows uploading it with
  `github/codeql-action/upload-sarif`. Other formats leave the output empty.
- Tests: every fixture gains `expected/sarif.golden`, and a unit test checks the log's
  required SARIF fields; no schema validator dependency.
- Acceptance: `task ci` green, every fixture passes in all four formats, and a run in a
  public test repo shows the findings in the Security tab after `upload-sarif`.

### 2026-10-27: releases with Release Please (accepted)

Problem: SPEC 11 had releases start from a pushed `v*` tag, with `CHANGELOG.md` written by
hand, while every commit already follows Conventional Commits.

Change: `.github/workflows/release.yml` runs Release Please on pushes to `main`. It keeps a
release PR open that bumps `.github/.release-please-manifest.json` and `CHANGELOG.md`;
merging it creates the `vX.Y.Z` tag and the GitHub Release. A Publish job in the same
workflow then runs goreleaser on the tag (`release.mode: keep-existing`, so the notes stay
Release Please's) and moves `v0` with `task release:float`, which skips pre-releases.
Publishing runs in the same workflow because a tag created with `GITHUB_TOKEN` starts no
workflows. Before 1.0, `feat` bumps the minor version, so the first release is v0.1.0.

The workflow uses `GITHUB_TOKEN`, not a GitHub App: the release PR therefore starts no CI
on its own, and is closed and reopened to run the required checks. goreleaser 2.18.3 is
pinned in `mise.toml` and the workflow, and the publish job uses no caches.

Two readings of SPEC 10 and 12, within their wording: the action's `working-directory` is
passed as `-C`, so annotation paths are relative to it and line up only when it is the
checked-out repository root (documented in the README); and the self-check job uses
`uses: ./` with an inline zizmor ignore, because zizmor prefers `$/` and actionlint 1.7.11
cannot parse it.

### 2026-10-27: README status table from CI badges (accepted)

Problem: the README showed no build status, and the per-check results lived only in the
pull request comment.

Change: a **Badges** job runs after Checks and Self-check on pushes to `main` and commits
one shields.io endpoint JSON per check to a `badges` branch through the Git Data API, so no
git credentials stay on the runner. `README.md` gets a Status section with Checks, Build and
tests, Security and Tech stack tables. shields.io reads the files from
`raw.githubusercontent.com`, which serves only public repositories, so the badges render
once the repo is public.

### 2026-10-27: PR check table, labels, assignment and a main ruleset (accepted)

Problem: CI ran `task ci` as one step, so a pull request showed one pass or fail with no
per-check results; pull requests got no labels or assignee; and nothing on GitHub
enforced the pull-request-only rule the git hooks apply locally.

Change: the `ci` job becomes **Checks**, one `task` step per check with `!cancelled()`,
and a **Check table** job comments each result on the pull request (words first, the
symbol only repeating it, per ACCESSIBILITY.md). `actions/labeler` labels by path and a
`github-script` step assigns the author; both are `uses:` steps, so the rule that every
`run:` step calls `task` holds. The `main` ruleset (no force pushes or deletion, linear
history, signed commits, squash-only pull requests, Checks and Self-check required) is set
in the repository settings, not kept in the tree. GitHub refuses rulesets and auto-merge on
a private repository without GitHub Pro, so it is applied when the repo goes public in
Phase 4.

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
