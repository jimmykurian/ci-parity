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

Out of scope for v0.1: Make, just and mise readers, unwrapping for suggestions and fix
mode (Phases 6 to 8 in `docs/PLAN.md`); PowerShell parsing; and the Backlog items,
reading `task --list-all --json` and a companion run-task action. SARIF output arrives in
v0.2 (Phase 5).

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
runner: taskfile                  # or make, just, mise (from v0.3; see 5.7)
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
- P001, from v0.3: both commands are also unwrapped once before comparing. A command
  that starts with a package runner below drops the runner words and the options after
  them; options are the arguments starting with `-` before the first other one, `--`
  ends them, `--flag=value` carries its value, and a listed flag also drops the
  argument after it. Nothing left means no unwrapping. An inline command and a task
  command match when they are equal as written or unwrapped, on either side. Only
  suggestions change.

  | Runner words | Options that take a separate value | Not unwrapped when |
  | --- | --- | --- |
  | `npx`, `pnpx`, `pnx`, `bunx` | `-p --package` | `-c` or `--call` is given |
  | `pnpm exec`, `pnpm dlx` | none | |
  | `uv run`, `uvx` | `--with --with-editable --with-requirements --python -p --project --directory --extra --group --package --env-file --from --index --default-index` | |
  | `dotnet tool run` | none | |

- P002: a task name that has the unknown name as a prefix wins, shortest first.
  Otherwise suggest the name with the lowest Levenshtein distance, when that distance is
  at most 3 and less than the length of the unknown name. Internal tasks are never
  suggested.

### 5.7 Other runners (from v0.3)

`runner` selects `taskfile`, `make`, `just` or `mise`. Each has a reader behind the same
interface and its own call syntax; the rules are unchanged.

- Discovery: the root file is found in the repository root, names matched exactly.
  make: `GNUmakefile`, `makefile`, `Makefile`. just: `justfile`, `Justfile`,
  `.justfile`. mise: every one of `.config/mise/config.toml`, `.config/mise.toml`,
  `mise/config.toml`, `.mise/config.toml`, `mise.toml` and `.mise.toml` that exists,
  read in that order, a later file's task replacing an earlier one's; messages name the
  last one found. The `taskfile` key is only for `runner: taskfile`; set with another
  runner it is a config error. A missing root file exits 2.
- Defaults by runner: `invokers.shell` is `[make, gmake]`, `[just]` or
  `["mise run", "mise r", "mise tasks run"]`, and `invokers.actions` is empty, unless
  the config sets them. Unless `tools.replace` is set, the configured runner's commands
  leave the tools list and the other runners' are on it, including `task` and `go-task`.
- make: a target is a rule line outside a recipe; special targets (`.PHONY`), pattern
  rules (`%`), variable assignments (including target-specific ones), `define` blocks
  and names holding `$` define none. `include`, `-include` and `sinclude` of literal
  paths, resolved from the repository root, are followed to depth 10; a missing file is
  skipped, as make may generate it. Recipe lines, without leading `@`, `-` or `+` and
  with whitespace collapsed, are a target's commands. No target is internal. In a call,
  `-f`, `--file`, `--makefile`, `-C` and `--directory` (in any form) select another
  makefile and the call is ignored; `-I --include-dir -o --old-file --assume-old -W
  --what-if --new-file --assume-new` take a value; `-j --jobs -l --load-average
  --max-load` take the next argument only when it is a number; `NAME=value` is a
  variable; a call naming no target runs the default goal and is ignored.
- just: recipes, their parameters, `alias a := b`, `import` of literal paths and `mod`
  (recipes named `mod::recipe`, from `name.just`, `name/mod.just`, `name/justfile` or a
  given path). A recipe is internal when it is `[private]` or its name starts with `_`.
  Body lines without leading `@` or `-` are its commands. In a call, the words after a
  recipe are its arguments, as many as its parameters take (all the rest for a variadic
  one); `just mod recipe` and `just mod::recipe` both name a module recipe. `-f`,
  `--justfile`, `-d` and `--working-directory` select another justfile; `--set` takes
  two values.
- mise: `[tasks.name]` tables and `name = "cmd"` (or list) entries, with `run` (string
  or list) as the commands, `alias` (string or list) as aliases and `hide = true` as
  internal; executable file tasks under `.config/mise/tasks/`, `mise/tasks/`,
  `.mise/tasks/`, `mise-tasks/` and `.mise-tasks/` (named from the path, `db/migrate`
  as `db:migrate`, hidden by a `#MISE hide=true` line, with no commands); and
  `task_config.includes`, each a task TOML file whose top-level tables are tasks or a
  directory of file tasks, a missing one skipped. A call names one task, then its
  arguments; `:::` starts another task. `-C` and `--cd` select another directory; a
  task name holding a glob (`lint:*`) is not checked; `-j --jobs -E --env -t --tool
  --timeout -o --output -s --shell` take a value.

## 6. Rules

| ID   | Name             | Severity                         | Trigger |
|------|------------------|----------------------------------|---------|
| P001 | inline-tool      | error, or tolerated via baseline, or allowed | A `run:` step calls a listed tool directly. |
| P002 | unknown-task     | error                            | An invoker names a task the Taskfile does not define. |
| P003 | internal-task    | error                            | An invoker names a task marked `internal: true`. |
| P004 | regression       | error                            | A workflow's P001 count exceeds its baseline, or a locked workflow has any P001. One finding per workflow, located at line 1 of the workflow. |
| P005 | unsupported      | warning                          | Unsupported shell, shell parse error, dynamic command name, allow comment without reason. |

Message wording (normative, because users grep for it). With another runner (5.7), its
command takes the place of `task` in `task <name>`: `make lint`, `just lint`,
`mise run lint`.

- P001: `step "<step>" in job "<job>" runs '<command>' directly` followed, when a
  suggestion exists, by `; already available as 'task <name>'`. For composite actions
  use `in action "<path>"` instead of `in job`.
- P002: `'task <name>' is not defined in <taskfile>` followed, when a suggestion exists,
  by ` (closest: <name>)`.
- P003: `'<name>' is marked internal: true and should not be invoked from CI`. With just
  it reads `is private`, and with mise `is hidden`, in place of `is marked internal: true`.
- P004: `<workflow> has <current> inline-tool findings, baseline allows <baseline>` or
  `<workflow> is locked at 0 and has <current> inline-tool findings`.
- P005: `step uses shell '<shell>'; not analyzed`, `composite step has no shell; not
  analyzed`, `cannot parse step script: <error>`,
  `dynamic command name '<word>' cannot be analyzed`, `allow comment has no reason`.

A step without a `name:` is referred to as `step <index>` (1-based) in place of
`step "<step>"`, unquoted: `step 4 in job "web" runs 'pnpm test' directly`.

## 7. CLI

```text
ci-parity check    [--config PATH] [--baseline PATH] [--format text|github|json|sarif]
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

SARIF format (`--format sarif`, from v0.2): one SARIF 2.1.0 log with one run, sorted like
the other formats.

- `runs[0].tool.driver`: `name` `ci-parity`, `version` from the build (plus
  `semanticVersion` when that version is `X.Y.Z`), `informationUri` the repository URL,
  and one `rules` entry per rule P001 to P005 with `id`, `name`, `shortDescription`,
  `helpUri` (the README's Rules section) and `defaultConfiguration.level`.
- One result per finding: `ruleId`, `level` `error` for errors, `warning` for warnings
  and `note` for tolerated findings, and `message.text` with the same message as the
  other formats. An allowed finding is a result with
  `suppressions: [{"kind": "inSource", "justification": <reason>}]`.
- `locations[0].physicalLocation`: `artifactLocation` with the repository-relative `uri`
  and `uriBaseId` `%SRCROOT%`; `region` with `startLine` and `startColumn`.
- `partialFingerprints["ciParity/v1"]`: a hash of the rule, file, job or action, step
  and command, so a finding keeps its identity when lines above it move.

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
- `format`: default `github`. With `sarif` the log is written to
  `$RUNNER_TEMP/ci-parity.sarif` instead of printed, and the `sarif-file` output holds
  its path; other formats leave that output empty.
- `strict`: default `false`.
- `working-directory`: default `.`.
- `fail-on-error`: default `true`. When `false`, the step succeeds and exposes outputs.

Outputs: `errors`, `warnings`, `tolerated`, read from a JSON run written to
`$RUNNER_TEMP`, and `sarif-file`.

Steps: resolve the version; download `ci-parity_<version>_<os>_<arch>.tar.gz` (zip on
windows) and `checksums.txt` from the release with `gh release download`, using
`GH_TOKEN: ${{ github.token }}`; verify with sha256; extract to `$RUNNER_TEMP/ci-parity`;
run `check` in `working-directory` once with `--format json` to a file for outputs and
once with the requested format for the log; honor `fail-on-error`.

Branding: `icon: git-merge`, `color: blue`.

## 11. Releases

goreleaser v2. Targets linux, darwin, windows on amd64 and arm64. Archive names
`ci-parity_<version>_<os>_<arch>.tar.gz`, zip on windows, plus `checksums.txt`.
Release Please runs on pushes to `main` and keeps a release PR that bumps the version and
`CHANGELOG.md` from the Conventional Commits; merging it creates the `vX.Y.Z` tag and the
GitHub Release. The same workflow then runs goreleaser on that tag to attach the archives,
and moves the floating major tag (`v0`) to the released commit so
`uses: OWNER/ci-parity@v0` tracks the latest v0.x. Version, commit and date are injected
with `-ldflags`.

## 12. This repo's own CI

`.github/workflows/ci.yml`:

- job `checks` ("Checks") on ubuntu: checkout, setup-go (reads `go.mod`), mise-action,
  gitleaks-action, then one `run: task <name>` step for each task in `task ci` except
  self-check, every step running even after one fails.
- job `self-check` ("Self-check"): checkout, setup-go, mise-action, `run: task build`, then
  `uses: ./` with `binary: ./bin/ci-parity`.
- job `report` ("Check table"), on pull requests: one comment with each step's result,
  updated on every push.
- job `badges` ("Badges"), on pushes to `main`: one shields.io endpoint file per check,
  committed to the `badges` branch, for the status table in `README.md`.

`.github/workflows/labeler.yml` labels pull requests by path (`.github/labeler.yml`), and
`.github/workflows/pr-assign-author.yml` assigns each new one to its author.
Once the repository is public, a ruleset on `main` requires squash-only pull requests,
signed commits, linear history, and Checks and Self-check.

`.github/workflows/release.yml` runs Release Please and goreleaser via `uses:`, and moves
the floating tag with `run: task release:float`.

The repo ships `.ci-parity.yml` and a `.ci-parity-baseline.json` with every workflow
locked at 0.

Dev tools (golangci-lint v2, govulncheck, actionlint, zizmor, gitleaks, shellcheck,
yamllint, markdownlint-cli2, typos, editorconfig-checker) are pinned in `mise.toml` and
never added to `go.mod`, as `tool` directives or otherwise; the dependency rule in
`AGENTS.md` covers `go.mod` and its four modules. In CI they come from a
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
