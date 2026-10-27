# ci-parity v0.1 specification

## 1. Purpose and positioning

ci-parity checks that GitHub Actions workflows invoke project commands through the
repository's task runner instead of calling build, test, lint and deploy tools inline.
It reports every inline invocation, verifies that every task CI references exists, and
keeps a per-workflow baseline so an existing repo can migrate one workflow at a time
without regressing.

The README opens with this positioning. Write it before writing code.

- actionlint checks workflow syntax, expression types and shellcheck over `run:` scripts.
  It has no notion of a task runner.
- zizmor reports security findings. No notion of a task runner.
- gha-workflow-linter validates that `uses:` references resolve to real refs.
- bitwarden/workflow-linter is an org-rule framework whose rules cannot span a workflow
  and a Taskfile.
- task2ci (archived June 2026) generated workflows from Taskfile annotations and failed
  CI when the generated file drifted. Generation only; it never linted hand-written
  workflows and had no baseline.

ci-parity is the enforcement approach: keep hand-written workflows, lint them, ratchet
the inline count down to zero.

## 2. Scope

In v0.1:

- Taskfile (go-task v3) as the only runner.
- GitHub Actions workflows and local composite actions.
- Steps whose shell is bash or sh.
- Rules P001 to P005.
- Output formats text, github, json.
- Baseline ratchet.
- Composite GitHub Action.
- Release binaries for linux, darwin, windows on amd64 and arm64.

Out of scope (Phase 5 backlog in `docs/PLAN.md`): Make, just and mise readers; SARIF;
auto-fix; PowerShell parsing; reading `task --list-all --json`; a companion run-task action.

## 3. Vocabulary

- Runner: the task runner. v0.1: Taskfile.
- Invoker: how CI calls the runner. A shell invoker (`task x`, `go-task x`,
  `go tool task x`) or an action invoker (a local composite action such as
  `./.github/actions/run-task` whose input names the task).
- Tool: a command on the tools list. Only tools are flagged; everything else is ignored.
- Finding: one rule hit at one location with severity `error`, `warning`, `tolerated`
  or `allowed`.
- Baseline: per-workflow count of tolerated P001 findings. Counts only go down. A count
  of zero, or a workflow missing from the baseline, locks that workflow.

## 4. Inputs

### 4.1 Files scanned

- `.github/workflows/*.yml` and `*.yaml`
- `.github/actions/**/action.yml` and `action.yaml` (local composite actions; their
  `runs.steps` are scanned like job steps)

Paths are reported relative to the repo root with forward slashes on all platforms.

Path globs (`workflows`, `allow[].workflows`) match repo-relative forward-slash paths.
`*` matches any characters except `/`, `?` matches one character except `/`, and a whole
`**` segment matches zero or more path segments. No other syntax (no braces, no
character classes). The standard library has no `**` support and no glob library is
allowed, so `internal/domain` implements this matcher itself, with table tests.

### 4.2 Config: `.ci-parity.yml`

Every key is optional. Defaults are shown.

```yaml
runner: taskfile
taskfile: ""                      # empty = auto-discover: Taskfile.yml, taskfile.yml,
                                  # Taskfile.yaml, taskfile.yaml, then the .dist variants
workflows:
  - .github/workflows/*.yml
  - .github/workflows/*.yaml
  - .github/actions/**/action.yml
  - .github/actions/**/action.yaml
invokers:
  shell: [task, go-task, "go tool task"]
  actions:
    - uses: ./.github/actions/run-task
      input: task
tools:
  extend: []                      # added to the built-in list
  remove: []                      # removed from the built-in list
  replace: []                     # non-empty replaces the built-in list entirely
transparent: [sudo, env, time, nice, ionice, timeout, nohup, xvfb-run, chronic]
allow: []                         # no commands are allowed by default
baseline: .ci-parity-baseline.json
```

An `allow` entry looks like this:

```yaml
allow:
  - match: "docker push *"        # glob matched against the normalized command
    reason: "images publish only from CI"   # required
    workflows: []                 # optional path globs; empty means every workflow
```

`allow[].match` is a command glob, not a path glob: it is matched against the whole
normalized command (5.3 step 6), `*` matches any run of characters including `/` and
spaces, and `?` matches exactly one character. `docker push *` therefore matches
`docker push ghcr.io/acme/widget-service:1.0.0`. Matching is case-sensitive.

Built-in tools, matched on the command's basename so `./node_modules/.bin/eslint` and
`./gradlew` match:

- node: `npm npx pnpm pnpx pnx yarn bun bunx node deno`
- python: `python python3 pip pip3 pipx uv uvx poetry pytest ruff mypy tox nox`
- dotnet: `dotnet`
- go: `go golangci-lint`
- rust: `cargo`
- jvm: `mvn mvnw gradle gradlew`
- infra: `terraform tofu docker docker-compose kubectl helm az aws gcloud pulumi`
- quality: `eslint prettier biome vitest jest playwright cypress k6 shellcheck actionlint`
- other runners, which count as bypasses when they are not the configured runner:
  `make just mise cargo-make`

Config errors (unknown key, wrong type, `allow` entry without `reason`, `replace` used
together with `extend` or `remove`) exit 2 with a message that names the key.

### 4.3 Baseline: `.ci-parity-baseline.json`

```json
{
  "version": 1,
  "workflows": {
    ".github/workflows/ci.yml": 7,
    ".github/workflows/release.yml": 0
  }
}
```

Semantics:

- No baseline file: strict mode. Every P001 finding is an error.
- Workflow present with count N greater than zero: the first N P001 findings in that
  workflow (file order, `allowed` findings excluded) are `tolerated`; the rest are
  `error`, and one P004 finding per workflow reports current versus baseline.
- Workflow present with count 0, or absent: locked. Every P001 finding is an error, and
  one P004 finding reports the lock.
- P002, P003 and P005 are never affected by the baseline.
- `ci-parity baseline` writes the current non-allowed P001 count for every scanned
  workflow, writing `0` explicitly so locks are visible in diffs. It refuses with exit 1
  if any count would rise, listing the workflows, unless `--allow-regression` is given.
- `ci-parity check` prints a ratchet table whenever a current count is below its
  baseline, suggesting `ci-parity baseline` to lock in the improvement. Not an error.
- `--strict` on `check` ignores the baseline entirely.

## 5. Analysis

### 5.1 Step enumeration

Decode each file with `yaml.v3` into a `yaml.Node` tree so every value has a line and
column. For workflows, walk `jobs.<id>.steps[]`. For composite actions, walk
`runs.steps[]`. Steps with `uses:` are checked against action invokers (5.4) and are
otherwise ignored. Steps with `run:` are analyzed.

For a `run:` scalar in literal or folded block style, the script's first line is the
node's line plus one; for a plain or quoted scalar it is the node's line. Report the
line of the command inside the script. Lines and columns are exact for plain, quoted,
literal and folded values. For a folded value, rebuild the folding from the source
lines to map each script position back to the source line it came from; if the rebuilt
text does not equal the parsed value, fall back to counting lines from the first one.
Positions found after the `${{ }}` replacement below are mapped back to the original
script text, so an expression earlier on the line does not shift the column.

On a plain or quoted single-line `run:` value, ` # ...` after the value is a YAML
comment that YAML strips from the value. Append it to the script before shell parsing
so `run: docker push acme # ci-parity: allow <reason>` works.

Replace every `${{ ... }}` expression in the script with the token `__GHA_EXPR_<n>__`
before shell parsing, where `<n>` is the expression index within the step.

### 5.2 Shell resolution

Resolution order: step `shell`, job `defaults.run.shell`, workflow `defaults.run.shell`,
then the OS default. The OS default is `pwsh` when the job's `runs-on` (string or list)
contains `windows`, otherwise `bash`. If `runs-on` contains an expression, assume `bash`.
Composite action steps must declare `shell`; read it. A composite `run:` step without
`shell` produces one P005 finding for the step and the step is skipped.

Analyze when the shell is `bash`, `sh`, or starts with `bash ` or `sh `. Any other shell
(`pwsh`, `powershell`, `cmd`, `python`, custom) produces one P005 finding for the step
and the step is skipped.

### 5.3 Command extraction

1. Parse the script with `mvdan.cc/sh/v3/syntax` using `syntax.LangBash` and
   `KeepComments(true)`. A parse error produces one P005 finding for the step, located
   at the error's position mapped back into the file, and the step is skipped. The
   message carries the parser's text without its script-relative position.
2. Walk the tree. For every `*syntax.CallExpr`, ignore its assignments and take the
   first argument. If the word is not a plain literal (it contains an expansion or an
   `__GHA_EXPR_` token), emit P005 "dynamic command name" and continue.
3. Peel transparent wrappers: while the command's basename is in `transparent`, drop it
   along with its options, then re-read the command. Options are arguments that start
   with `-` or contain `=`. A flag listed below for that wrapper also drops the argument
   after it; `--flag=value` and attached short forms (`-uroot`) carry their own value.
   For `timeout` and `nice`, one numeric argument after the options (`30`, `30s`) is
   dropped too. Wrappers added through config get the generic rule only.

   | Wrapper    | Flags that take a separate value |
   |------------|----------------------------------|
   | `sudo`     | `-u -g -h -p -C -D -R -r -t -T -U --user --group --host --prompt --close-from --chdir --chroot --role --type --command-timeout --other-user` |
   | `env`      | `-u -C -S --unset --chdir --split-string` |
   | `time`     | `-f -o --format --output` |
   | `nice`     | `-n --adjustment` |
   | `ionice`   | `-c -n -p -P -u --class --classdata --pid --pgid --uid` |
   | `timeout`  | `-k -s --kill-after --signal` |
   | `xvfb-run` | `-n -s -f -e -p -w --server-num --server-args --auth-file --error-file --xauth-protocol --wait` |

   `env -S`/`--split-string` hides the command inside a string, so the call is dropped
   once its value is consumed, as is a wrapper left with no command (`sudo -v`).
4. Take the basename of the command.
5. If it begins an invoker pattern (5.4), extract task names and continue to rule
   evaluation for P002 and P003.
6. Else if it is on the tools list, record a P001 candidate with the normalized command
   (words joined by single spaces, literal words only).
7. Else ignore.

Calls inside `if`, `for`, `while`, `case`, pipelines, `&&`/`||` lists, subshells and
function bodies are all walked. Heredoc bodies are not analyzed.

Allow comments: a comment of the form `# ci-parity: allow <reason>` on the same line as
the command or on the line directly above it marks that command `allowed`. A missing
reason emits P005 "allow comment without reason" and the command is not allowed.

### 5.4 Invoker detection

Shell invokers are word sequences. `go tool task` matches the three words `go`, `tool`,
`task`. After the invoker words, parse arguments:

- Stop at `--`.
- Skip flags. These flags consume the next argument: `-t --taskfile -d --dir
  -C --concurrency -I --interval -o --output`. A flag written as `--flag=value`
  carries its value and consumes nothing further. Any other argument starting with `-`
  is a flag that consumes nothing.
- Arguments of the form `KEY=VALUE` are task variables, not task names.
- `-t`, `--taskfile`, `-d`, `--dir` (in any form, including `--dir=web` and `-dweb`),
  `-g` and `--global` select a Taskfile other than the root one, so the call is ignored.
- Every remaining argument is a task name. A task name that is not a literal word cannot
  be checked and is skipped. Zero task names (for example `task --list`) means the call
  is ignored.

Action invokers: a `uses:` step matches when its value equals the configured `uses`
exactly, or equals it followed by `@<ref>`. Task names come from `with.<input>`, split on
whitespace; a word that holds `${{ }}` cannot be checked and is skipped. A matching step
without that input is ignored.

### 5.5 Taskfile reading

Static parsing of Taskfile v3 (a missing `version` is treated as 3; any other version
exits 2). Collect:

- Task names from `tasks`, plus every entry of each task's `aliases`.
- `internal: true` per task.
- `includes`: a string value is a path to a Taskfile or to a directory (the first of the
  root discovery names found there); a map value supports `taskfile`, `dir`, `optional`, `internal`,
  `aliases`, `flatten`. Included tasks are named `<namespace>:<task>` (also under each
  alias namespace), or unprefixed when `flatten: true`. `internal: true` on an include
  marks all its tasks internal. Resolve relative to the including file's directory, cap
  recursion at depth 10, ignore cycles. A missing include is a parse error (exit 2)
  unless `optional: true`. An include path holding a template (`{{OS}}`) is treated as
  missing.
- An included Taskfile with a `default` task also makes the bare namespace a valid name.
- File names are matched exactly against the directory listing.
- Each task's command strings for suggestion matching: `cmd:` on the task, string
  entries of `cmds`, and `cmd:` map entries of `cmds`. Ignore `task:`, `for:` and
  `defer:` entries.

### 5.6 Suggestions

- P001: normalize each Taskfile command string (trim, collapse whitespace). If it equals
  the inline command's normalized form, suggest `task <name>`, never suggesting internal
  tasks and choosing the shortest name. Ties go to the shorter name, then the
  alphabetically first. Only the canonical task name is suggested, not an alias. Treat
  `pnpm run X` as equal to `pnpm X` and `yarn run X` as equal to `yarn X`.
- P002: a task name that has the unknown name as a prefix wins, shortest first.
  Otherwise suggest the name with the lowest Levenshtein distance, when that distance is
  at most 3 and less than the length of the unknown name. Internal tasks are never
  suggested.

## 6. Rules

| ID   | Name             | Severity                         | Trigger |
|------|------------------|----------------------------------|---------|
| P001 | inline-tool      | error, or tolerated via baseline, or allowed | A `run:` step calls a listed tool directly. |
| P002 | unknown-task     | error                            | An invoker names a task the Taskfile does not define. |
| P003 | internal-task    | error                            | An invoker names a task marked `internal: true`. |
| P004 | regression       | error                            | A workflow's P001 count exceeds its baseline, or a locked workflow has any P001. One finding per workflow, located at line 1 of the workflow. |
| P005 | unsupported      | warning                          | Unsupported shell, shell parse error, dynamic command name, allow comment without reason. |

Message wording (normative, because users grep for it):

- P001: `step "<step>" in job "<job>" runs '<command>' directly` followed, when a
  suggestion exists, by `; already available as 'task <name>'`. For composite actions
  use `in action "<path>"` instead of `in job`.
- P002: `'task <name>' is not defined in <taskfile>` followed, when a suggestion exists,
  by ` (closest: <name>)`.
- P003: `'<name>' is marked internal: true and should not be invoked from CI`.
- P004: `<workflow> has <current> inline-tool findings, baseline allows <baseline>` or
  `<workflow> is locked at 0 and has <current> inline-tool findings`.
- P005: `step uses shell '<shell>'; not analyzed`, `composite step has no shell; not
  analyzed`, `cannot parse step script: <error>`,
  `dynamic command name '<word>' cannot be analyzed`, `allow comment has no reason`.

A step without a `name:` is referred to as `step <index>` (1-based) in place of
`step "<step>"`, unquoted: `step 4 in job "web" runs 'pnpm test' directly`.

## 7. CLI

```text
ci-parity check    [--config PATH] [--baseline PATH] [--format text|github|json]
                   [--strict] [--verbose] [-C DIR]
ci-parity baseline [--config PATH] [--baseline PATH] [--allow-regression] [-C DIR]
ci-parity version
```

- `--config` and `--baseline` override the config path and the baseline path from the
  config. `-C` changes to DIR before doing anything.
- `--verbose` includes tolerated and allowed findings in text output. JSON always
  includes them.
- `version` prints `ci-parity <version> (<commit>, <date>)`, with values injected by
  the build via `-ldflags`.

## 8. Output formats

Findings are sorted by file, line, column, rule ID. Text format:

```text
.github/workflows/ci.yml
  42:9   P001 inline-tool   step "Run tests" in job "test" runs 'pnpm test' directly; already available as 'task test'
  58:9   P002 unknown-task  'task lint:all' is not defined in Taskfile.yml (closest: lint:ci)

ratchet
  .github/workflows/ci.yml       baseline 7   now 5   run `ci-parity baseline` to lock in
  .github/workflows/release.yml  locked       now 1   REGRESSION

2 errors, 0 warnings, 5 tolerated, 1 allowed
```

The ratchet block is omitted when there is nothing to report.

Text format details:

- Position and rule name are padded into columns: the position column is as wide as
  the longest `line:col` among the lines shown plus three spaces, and the name column
  is 13 characters (`internal-task`) plus one space.
- Files with no lines shown are omitted, and file groups are separated by a blank line.
- The summary line is always printed, also when there are no findings. `error` and
  `warning` are singular for a count of one (`1 error, 2 warnings`); `tolerated` and
  `allowed` never change.
- With `--verbose`, tolerated lines end in ` [tolerated]` and allowed lines end in
  ` [allowed: <reason>]`.

GitHub format emits one workflow command per finding, nothing else:

```text
::error file=.github/workflows/ci.yml,line=42,col=9,title=P001 inline-tool::step "Run tests" in job "test" runs 'pnpm test' directly; already available as 'task test'
::warning file=...,title=P005 unsupported::step uses shell 'pwsh'; not analyzed
::notice file=...,title=P001 inline-tool (tolerated)::...
```

Errors use `::error`, warnings `::warning`, tolerated `::notice`, allowed nothing.
Messages are escaped per the workflow-command rules (`%`, `\r`, `\n`, `:`, `,`).

JSON format:

```json
{
  "version": 1,
  "summary": { "errors": 2, "warnings": 0, "tolerated": 5, "allowed": 1 },
  "findings": [
    {
      "rule": "P001",
      "name": "inline-tool",
      "severity": "error",
      "file": ".github/workflows/ci.yml",
      "line": 42,
      "col": 9,
      "job": "test",
      "step": "Run tests",
      "command": "pnpm test",
      "message": "step \"Run tests\" in job \"test\" runs 'pnpm test' directly; already available as 'task test'",
      "suggestion": "task test",
      "reason": null
    }
  ],
  "ratchet": {
    ".github/workflows/ci.yml": { "baseline": 7, "current": 5, "locked": false }
  }
}
```

`job` is null for composite actions and `action` carries the path instead. `reason` is
the allow reason for allowed findings and null otherwise. The schema is documented in
the README and only changes with a `version` bump.

## 9. Exit codes

- 0: no error-severity findings.
- 1: at least one error-severity finding, or `baseline` refused a regression.
- 2: usage error, config error, unreadable or unparseable YAML, missing Taskfile.

## 10. GitHub Action

`action.yml` at the repo root (required for the Marketplace), `runs.using: composite`.

Inputs:

- `version`: release tag to run. Default empty, which resolves to `github.action_ref`
  when that ref is an exact `vX.Y.Z` tag, otherwise to the latest release.
- `binary`: path to a prebuilt `ci-parity`. When set, no download happens. This repo's
  own CI uses it to test the action before a release exists.
- `config`, `baseline`: passed through.
- `format`: default `github`.
- `strict`: default `false`.
- `working-directory`: default `.`.
- `fail-on-error`: default `true`. When `false`, the step succeeds and exposes outputs.

Outputs: `errors`, `warnings`, `tolerated`, read from a JSON run written to
`$RUNNER_TEMP`.

Steps: resolve the version; download `ci-parity_<version>_<os>_<arch>.tar.gz` (zip on
windows) and `checksums.txt` from the release with `gh release download`, using
`GH_TOKEN: ${{ github.token }}`; verify with sha256; extract to `$RUNNER_TEMP/ci-parity`;
run `check` in `working-directory` once with `--format json` to a file for outputs and
once with the requested format for the log; honor `fail-on-error`.

Branding: `icon: git-merge`, `color: blue`.

## 11. Releases

goreleaser v2. Targets linux, darwin, windows on amd64 and arm64. Archive names
`ci-parity_<version>_<os>_<arch>.tar.gz`, zip on windows, plus `checksums.txt`.
A release workflow runs on tags matching `v*`, publishes the GitHub Release, and moves
the floating major tag (`v0`) to the released commit so `uses: OWNER/ci-parity@v0`
tracks the latest v0.x. Version, commit and date are injected with `-ldflags`.

## 12. This repo's own CI

`.github/workflows/ci.yml`:

- job `checks` ("Checks") on ubuntu: checkout, setup-go (reads `go.mod`), mise-action,
  gitleaks-action, then one `run: task <name>` step for each task in `task ci` except
  self-check, every step running even after one fails.
- job `self-check` ("Self-check"): checkout, setup-go, mise-action, `run: task build`, then
  `uses: ./` with `binary: ./bin/ci-parity` (Phase 4; before that, `run: task self-check`).
- job `report` ("Check table"), on pull requests: one comment with each step's result,
  updated on every push.

`.github/workflows/labeler.yml` labels pull requests by path (`.github/labeler.yml`), and
`.github/workflows/pr-assign-author.yml` assigns each new one to its author.
Once the repository is public, a ruleset on `main` requires squash-only pull requests,
signed commits, linear history, and Checks and Self-check.

`.github/workflows/release.yml` runs goreleaser via `uses:` and moves the floating tag
with `gh`; those `run:` lines carry `# ci-parity: allow` comments with reasons.

The repo ships `.ci-parity.yml` and a `.ci-parity-baseline.json` with every workflow
locked at 0.

Dev tools (golangci-lint v2, govulncheck, actionlint, zizmor, gitleaks, shellcheck,
yamllint, markdownlint-cli2, typos, editorconfig-checker) are pinned in `mise.toml` and
never added to `go.mod`, as `tool` directives or otherwise; the dependency rule in
`AGENTS.md` covers `go.mod` and stays at three modules. In CI they come from a
SHA-pinned `jdx/mise-action` step before the `task` steps; gitleaks also runs as its own
`uses:` step, not part of `task ci`. Locally `mise install` sets them up; each
`task lint:<group>` checks its binaries are on `PATH` and fails with an install hint
naming the missing one. `.golangci.yml` uses the v2 format (`version: "2"`).

gitleaks-action needs a `GITLEAKS_LICENSE` secret only for repos owned by an
organization. For a personal-account repo it runs with `GITHUB_TOKEN` alone.

## 13. Non-functional

- A repo with 50 workflows checks in well under a second.
- No network access at check time.
- Deterministic output across platforms.
- Clear error for a config key that does not exist (exit 2, name the key).
