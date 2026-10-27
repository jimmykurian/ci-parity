# Build plan

Work one phase at a time. A phase is done when every checkbox is ticked, the acceptance
criteria pass, and `task ci` is green. Then commit, summarize, and stop. Do not start
the next phase until asked. Phase 5 is a backlog, not work.

Fixture layout, used from Phase 1 on:

```
testdata/fixtures/<case>/
  .ci-parity.yml                 (optional; defaults apply when absent)
  .ci-parity-baseline.json       (optional)
  Taskfile.yml                   (plus any included Taskfiles)
  .github/workflows/*.yml
  .github/actions/**/action.yml  (optional)
  expected/text.golden
  expected/github.golden
  expected/json.golden
  expected/exitcode              (single integer)
```

## Phase 0: scaffold

- [x] `go mod init github.com/OWNER/ci-parity`; OWNER from the git remote, ask if absent
- [x] `LICENSE` (Apache-2.0); `.gitignore`, `.gitattributes` and `.editorconfig` already exist, extend them if needed
- [x] `README.md`: positioning paragraph from SPEC section 1, a "pre-release" notice,
      and empty headings for Install, Quick start, Configuration, Rules, Baseline, FAQ
- [x] `cmd/ci-parity/main.go` with subcommand dispatch and a working `version`;
      `check` and `baseline` exit 2 with "not implemented"
- [x] `Taskfile.yml` with `build`, `test`, `lint`, `fmt`, `fmt:check`, `self-check`,
      `scrub`, `ci`. `self-check` runs `go run ./cmd/ci-parity version` for now.
      `scrub` warns and passes when `$SCRUB_TERMS_FILE` is unset
- [x] `.golangci.yml` in golangci-lint v2 format (`version: "2"`): standard linters
      plus `gocritic` and `revive`
- [x] Dev tools per SPEC section 12: nothing added to `go.mod`; `task lint` fails with an
      install hint when golangci-lint or actionlint is missing from `PATH`
- [x] `.github/workflows/ci.yml` per SPEC section 12 with `run: task ci` and
      `run: task self-check`, third-party actions pinned to SHAs, golangci-lint and
      actionlint installed by `uses:` steps, gitleaks as its own `uses:` step
- [x] `.ci-parity.yml` and `.ci-parity-baseline.json` for this repo (workflows locked at 0)
- [x] `internal/testutil` golden harness: discovers fixtures, runs the CLI in-process in
      each format, compares to goldens, `-update` flag rewrites them. Zero fixtures yet
- [x] `docs/DECISIONS.md` created with a short header

Acceptance: `task ci` green locally and in CI. `go run ./cmd/ci-parity version` prints a
version string.

## Phase 1: workflows, shell extraction, P001, text output

- [x] `internal/config`: load `.ci-parity.yml`, apply defaults, validate per SPEC 4.2,
      unknown keys are errors
- [x] Path glob matcher with `**` per SPEC 4.1 and command glob for `allow.match` per
      SPEC 4.2, both in `internal/config` with table tests (`**` at start, middle, end,
      zero segments; `*` across `/` in commands only)
- [x] `internal/workflow`: enumerate steps with positions per SPEC 5.1 and 5.2,
      including composite actions and composite steps without `shell` (P005)
- [x] `internal/shell`: command extraction per SPEC 5.3, including transparent wrappers,
      `${{ }}` replacement, allow comments
- [x] `internal/rules`: P001 and P005; a `Finding` type shared by every reporter
- [x] `internal/report/text`
- [x] `check --format text` wired end to end; `self-check` switched to `check`
- [x] Fixtures: `clean`, `inline-tools`, `wrappers`, `allow-comment`, `allow-config`,
      `dynamic-command`, `pwsh-step`, `composite-action-steps`, `composite-no-shell`,
      `nested-actions-glob`, `block-scalar-lines`
- [x] Unit tests for the extractor: pipelines, `&&` lists, `if` blocks, functions,
      heredocs, `cd dir && tool`, `env FOO=1 tool`, `timeout 30 tool`

Acceptance: all fixtures pass in text format. `task self-check` passes on this repo.
Reported line numbers are exact for both plain and block-scalar `run:` values.

## Phase 2: Taskfile, invokers, P002 and P003, suggestions

- [x] `internal/runner`: `Runner` interface (`Tasks()`, `IsInternal(name)`,
      `Commands(name)`) so Make, just and mise can be added later without touching rules
- [x] `internal/runner/taskfile`: static reader per SPEC 5.5 with includes, aliases,
      flatten, internal, optional
- [x] Shell invoker parsing per SPEC 5.4, including `go tool task`, flag values in both
      `--flag value` and `--flag=value` forms
- [x] Action invoker matching per SPEC 5.4
- [x] P002 and P003
- [x] Suggestions per SPEC 5.6 (P001 "already available as", P002 "closest")
- [x] Fixtures: `includes-nested`, `includes-optional-missing`, `internal-task`,
      `unknown-task`, `composite-invoker`, `go-tool-task`, `suggestion-match`,
      `task-vars-and-flags`
- [x] Unit tests for the Taskfile reader with three levels of includes and a cycle

Acceptance: all fixtures pass. A Taskfile with nested includes resolves namespaced
names. Unknown-task suggestions appear only within the distance rule.

## Phase 3: baseline, ratchet, formats, exit codes

- [ ] `internal/baseline`: read, write, compare per SPEC 4.3
- [ ] P004 and the tolerated severity
- [ ] `baseline` subcommand with `--allow-regression`
- [ ] `check --strict`, `check --verbose`
- [ ] `internal/report/github` with workflow-command escaping
- [ ] `internal/report/json` per SPEC 8, with a schema section in the README
- [ ] Exit codes per SPEC 9
- [ ] Fixtures: `baseline-within`, `baseline-regression`, `baseline-locked`,
      `baseline-improved`, `baseline-missing-strict`; github and json goldens for every
      existing fixture
- [ ] Harness asserts exit codes from `expected/exitcode`

Acceptance: every fixture passes in all three formats with the expected exit code.
`ci-parity baseline` on `baseline-regression` refuses without `--allow-regression`.

## Phase 4: action, releases, docs, v0.1.0

- [ ] `action.yml` per SPEC 10, including the `binary` input and outputs
- [ ] `.goreleaser.yaml` per SPEC 11 with `-ldflags` for version, commit, date
- [ ] `.github/workflows/release.yml`: goreleaser on `v*` tags, floating `v0` tag move,
      allow comments on the non-task `run:` lines
- [ ] `self-check` job switched to `uses: ./` with `binary: ./bin/ci-parity`
- [ ] README complete: install (release binary, `go install`, action), quick start with
      a before/after workflow snippet, configuration reference, rules reference, baseline
      workflow, JSON schema, FAQ (why not actionlint or zizmor, what about task2ci, what
      about Make or just, how to allow a step)
- [ ] `CHANGELOG.md` with a v0.1.0 entry
- [ ] Marketplace checklist verified: public repo, root `action.yml` with `name`,
      `description`, `branding`, a published release
- [ ] Tag `v0.1.0`

Acceptance: a throwaway repo with one inline `pnpm test` step and
`uses: OWNER/ci-parity@v0` gets an inline annotation on the offending line. Release
assets and `checksums.txt` are attached to the release. `task ci` green.

## Phase 5: backlog (do not start)

- Make, just and mise readers behind the `Runner` interface
- SARIF output for code scanning
- `task --list-all --json` as an optional source when the binary is present
- `fix` mode that rewrites a flagged step to its suggested task
- Companion `run-task` composite action in this repo
- Unwrapping `npx X`, `pnpm exec X`, `uv run X`, `dotnet tool run X` for suggestions
