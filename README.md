<h1>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.svg">
    <img src="docs/assets/logo-light.svg" alt="ci-parity" width="420" height="96">
  </picture>
</h1>

> **Pre-1.0.** ci-parity follows semantic versioning, and until v1.0.0 a minor release
> may change commands, config keys and output. Rule IDs, the baseline format and the JSON
> schema change only with a new `version` in the file.

ci-parity checks that GitHub Actions workflows invoke project commands through the
repository's task runner instead of calling build, test, lint and deploy tools inline.
It reports every inline invocation, verifies that every task CI references exists, and
keeps a per-workflow baseline so an existing repo can migrate one workflow at a time
without regressing.

## Status

Each badge is a check from the last CI run on `main` (`.github/workflows/ci.yml` writes them to the `badges` branch). On a pull request, the same checks report in a comment on the PR.

### Checks

| Branch | Format | Lint | Workflows | Shell | YAML | Markdown | Spelling | EditorConfig | Release config |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| main | [![Format](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Fformat.json&style=flat-square&logo=go&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Lint](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Flint.json&style=flat-square&logo=go&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Workflows](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Fworkflows.json&style=flat-square&logo=githubactions&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Shell](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Fshell.json&style=flat-square&logo=gnubash&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![YAML](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Fyaml.json&style=flat-square&logo=yaml&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Markdown](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Fmarkdown.json&style=flat-square&logo=markdown&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Spelling](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Fspelling.json&style=flat-square&logo=grammarly&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![EditorConfig](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Feditorconfig.json&style=flat-square&logo=editorconfig&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Release config](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fchecks%2Frelease.json&style=flat-square&logo=goreleaser&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) |

### Build and tests

| Branch | Tests | Build | Self-check |
| --- | --- | --- | --- |
| main | [![Tests](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fbuild%2Ftests.json&style=flat-square&logo=go&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Build](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fbuild%2Fbuild.json&style=flat-square&logo=go&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Self-check](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fbuild%2Fself-check.json&style=flat-square&logo=githubactions&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) |

### Security

| Branch | gitleaks | govulncheck | Scrub |
| --- | --- | --- | --- |
| main | [![gitleaks](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fsecurity%2Fgitleaks.json&style=flat-square&logo=git&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![govulncheck](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fsecurity%2Fgovulncheck.json&style=flat-square&logo=go&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) | [![Scrub](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Fjimmykurian%2Fci-parity%2Fbadges%2Fsecurity%2Fscrub.json&style=flat-square&logo=git&logoColor=white&label=%20&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions/workflows/ci.yml?query=branch%3Amain) |

### Tech stack

| Go | Task | mise | golangci-lint | yaml.v3 | mvdan/sh | GitHub Actions |
| --- | --- | --- | --- | --- | --- | --- |
| [![Go](https://img.shields.io/badge/%20-1.27-00ADD8?style=flat-square&logo=go&logoColor=white&labelColor=555)](https://go.dev/) | [![Task](https://img.shields.io/badge/%20-3-29BEB0?style=flat-square&logo=task&logoColor=white&labelColor=555)](https://taskfile.dev/) | [![mise](https://img.shields.io/badge/%20-tool%20versions-00ADD8?style=flat-square&logo=gnometerminal&logoColor=white&labelColor=555)](https://mise.jdx.dev/) | [![golangci-lint](https://img.shields.io/badge/%20-2.15-00ADD8?style=flat-square&logo=go&logoColor=white&labelColor=555)](https://golangci-lint.run/) | [![yaml.v3](https://img.shields.io/badge/%20-3.0-CB171E?style=flat-square&logo=yaml&logoColor=white&labelColor=555)](https://github.com/go-yaml/yaml) | [![mvdan/sh](https://img.shields.io/badge/%20-3.14-4EAA25?style=flat-square&logo=gnubash&logoColor=white&labelColor=555)](https://github.com/mvdan/sh) | [![GitHub Actions](https://img.shields.io/badge/%20-CI%20and%20Action-2088FF?style=flat-square&logo=githubactions&logoColor=white&labelColor=555)](https://github.com/jimmykurian/ci-parity/actions) |

How it differs from nearby tools:

- **actionlint** checks workflow syntax, expression types and shellcheck over `run:`
  scripts. It has no notion of a task runner.
- **zizmor** reports security findings. No notion of a task runner.
- **gha-workflow-linter** validates that `uses:` references resolve to real refs.
- **bitwarden/workflow-linter** is an org-rule framework whose rules cannot span a
  workflow and a Taskfile.
- **task2ci** (archived June 2026) generated workflows from Taskfile annotations and
  failed CI when the generated file drifted. Generation only; it never linted
  hand-written workflows and had no baseline.

ci-parity is the enforcement approach: keep hand-written workflows, lint them, ratchet
the inline count down to zero.

## Install

### GitHub Action

```yaml
- uses: actions/checkout@v7
- uses: jimmykurian/ci-parity@v0
```

`@v0` follows the latest v0.x release. To pin, use a release tag (`@v0.1.0`) or, better,
its full commit SHA with the tag as a comment. The action downloads that release's binary
for the runner, verifies it against `checksums.txt`, and runs `ci-parity check`.

| Input | Default | Meaning |
| --- | --- | --- |
| `version` | the action's ref | Release tag to run. Empty uses the action's own ref when it's an exact `vX.Y.Z` tag, otherwise the latest release. |
| `binary` | | Path to a prebuilt `ci-parity`; nothing is downloaded. |
| `config` | `.ci-parity.yml` | Passed to `--config`. |
| `baseline` | from the config | Passed to `--baseline`. |
| `format` | `github` | `github` (annotations on the changed lines), `text` or `json` for the log, or `sarif` for a file; see [SARIF output](#sarif-output). |
| `strict` | `false` | Ignore the baseline. |
| `working-directory` | `.` | Repository root to check (`-C`). Annotation paths are relative to it, so they line up only when it's the root of the checked-out repository. |
| `fail-on-error` | `true` | When `false`, the step succeeds and the outputs report the counts. |

| Output | Meaning |
| --- | --- |
| `errors` | Number of error findings. |
| `warnings` | Number of warning findings. |
| `tolerated` | Number of findings the baseline tolerates. |
| `sarif-file` | Path of the SARIF log when `format` is `sarif`; empty otherwise. |

### Release binaries

Each [release](https://github.com/jimmykurian/ci-parity/releases) has archives for
linux, darwin and windows on amd64 and arm64, and a `checksums.txt`:

```sh
gh release download v0.1.0 --repo jimmykurian/ci-parity \
  --pattern 'ci-parity_0.1.0_linux_amd64.tar.gz' --pattern checksums.txt
sha256sum --ignore-missing -c checksums.txt
tar -xzf ci-parity_0.1.0_linux_amd64.tar.gz
./ci-parity version
```

### go install

```sh
go install github.com/jimmykurian/ci-parity/cmd/ci-parity@latest
```

A `go install` build reports its version as `dev`; the release binaries carry the
version, commit and date.

## Quick start

A workflow that calls pnpm directly, next to a Taskfile that already wraps those commands:

```yaml
# .github/workflows/ci.yml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - name: Install
        run: pnpm install --frozen-lockfile
      - name: Test
        run: pnpm test
      - name: Lint
        run: task lint
```

```yaml
# Taskfile.yml
version: "3"

tasks:
  install:
    cmds:
      - pnpm install --frozen-lockfile
  test:
    cmds:
      - pnpm test
```

```text
$ ci-parity check
.github/workflows/ci.yml
  9:14    P001 inline-tool   step "Install" in job "test" runs 'pnpm install --frozen-lockfile' directly; already available as 'task install'
  11:14   P001 inline-tool   step "Test" in job "test" runs 'pnpm test' directly; already available as 'task test'
  13:19   P002 unknown-task  'task lint' is not defined in Taskfile.yml (closest: test)

3 errors, 0 warnings, 0 tolerated, 0 allowed
```

After calling the tasks instead (and installing task in a `uses:` step), it's clean:

```yaml
    steps:
      - uses: actions/checkout@v7
      - uses: go-task/setup-task@v2
      - name: Install
        run: task install
      - name: Test
        run: task test
```

```text
$ ci-parity check
0 errors, 0 warnings, 0 tolerated, 0 allowed
```

For a repository with many inline steps, start from a [baseline](#baseline) instead of
fixing everything at once:

```sh
ci-parity baseline     # records today's counts in .ci-parity-baseline.json; commit it
```

## Configuration

`.ci-parity.yml` at the repository root is optional, and every key is optional. These are
the defaults:

```yaml
runner: taskfile                  # or make or just; see Other runners
taskfile: ""                      # empty: Taskfile.yml, taskfile.yml, Taskfile.yaml,
                                  # taskfile.yaml, then the .dist variants
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
  replace: []                     # non-empty replaces the built-in list
transparent: [sudo, env, time, nice, ionice, timeout, nohup, xvfb-run, chronic]
allow: []
baseline: .ci-parity-baseline.json
```

- **Tools** are matched on the command's basename, so `./node_modules/.bin/eslint` and
  `./gradlew` count. The built-in list covers node (`npm npx pnpm pnpx pnx yarn bun bunx
  node deno`), python (`python python3 pip pip3 pipx uv uvx poetry pytest ruff mypy tox
  nox`), `dotnet`, go (`go golangci-lint`), `cargo`, jvm (`mvn mvnw gradle gradlew`),
  infrastructure (`terraform tofu docker docker-compose kubectl helm az aws gcloud
  pulumi`), quality tools (`eslint prettier biome vitest jest playwright cypress k6
  shellcheck actionlint`) and other task runners (`make just mise cargo-make`, plus `task`
  and `go-task` when the runner isn't Taskfile). The configured runner is never on the
  list. Anything else is ignored.
- **Transparent** wrappers are looked through: `sudo -E env CI=1 pnpm test` is `pnpm test`.
- **Invokers** are how CI calls the task runner: shell commands, or a local action whose
  input names the task.
- **Workflow globs** match repository-relative paths: `*` and `?` stay within one path
  segment, and a whole `**` segment matches any number of them.

### Other runners

With `runner: make`, ci-parity reads the `GNUmakefile`, `makefile` or `Makefile` in the
repository root instead of a Taskfile, with its `include`s. CI calls targets with `make`
or `gmake`, messages and suggestions say `make <target>`, and `task` counts as an inline
tool. A call with `-f` or `-C` points at another makefile and isn't checked, and plain
`make` (the default goal) isn't either. `.PHONY`, pattern rules and variables aren't
targets.

With `runner: just`, ci-parity reads the `justfile` (or `Justfile`, `.justfile`) with its
`import`s and `mod`ules. The words after a recipe are its arguments, as many as its
parameters take, so `just deploy staging` checks only `deploy`; module recipes are named
`db::migrate`, called as `just db migrate` or `just db::migrate`. A `[private]` recipe, or
one whose name starts with `_`, is reported as P003. `--list`, `--show` and other
subcommands, and calls with `-f` or `-d`, aren't checked. mise is coming next.

An unknown key, a wrong type, an `allow` entry without a `reason`, or `replace` combined
with `extend` or `remove` is an error that names the key (exit 2).

### Allowing a step

Some inline commands are deliberate. Allow one with a comment on the same line or the
line above, with a reason:

```yaml
      - run: docker push ghcr.io/acme/widget-service:dev # ci-parity: allow images publish only from CI
```

Or allow a command pattern in the config, optionally for some workflows only:

```yaml
allow:
  - match: "docker push *"           # matched against the whole command
    reason: images publish only from CI
    workflows: [.github/workflows/release.yml]
```

Allowed findings don't fail the check; `--verbose` lists them with their reason.

## Rules

| ID | Name | Severity | Reported when |
| --- | --- | --- | --- |
| P001 | inline-tool | error, tolerated or allowed | A `run:` step calls a listed tool directly. |
| P002 | unknown-task | error | CI calls a task the Taskfile doesn't define. |
| P003 | internal-task | error | CI calls a task marked `internal: true`. |
| P004 | regression | error | A workflow has more inline-tool findings than its baseline allows, or any while locked at 0. |
| P005 | unsupported | warning | A step ci-parity can't analyze: a shell other than bash or sh, a script that doesn't parse, a dynamic command name, or an allow comment without a reason. |

A P001 suggestion appears when a task runs the same command. Package runners are looked
through on both sides: `npx eslint .` suggests a task that runs `eslint .`, and an inline
`pytest` suggests a task that runs `uv run pytest`. The runners are `npx`, `pnpx`, `pnx`,
`bunx`, `pnpm exec`, `pnpm dlx`, `uv run`, `uvx` and `dotnet tool run`; `pnpm run X` and
`yarn run X` also match `pnpm X` and `yarn X`.

The messages are stable, so you can search for them:

- P001: `step "Test" in job "test" runs 'pnpm test' directly; already available as 'task test'`
- P002: `'task lint' is not defined in Taskfile.yml (closest: test)`
- P003: `'deploy:prod' is marked internal: true and should not be invoked from CI` (with
  just, `is private`)
- P004: `.github/workflows/ci.yml has 3 inline-tool findings, baseline allows 2`
- P005: `step uses shell 'pwsh'; not analyzed`

Exit codes: `0` when there are no errors, `1` when there are (or `baseline` refused to
raise a count), `2` for a usage or config error, YAML that doesn't parse, or a missing
Taskfile.

## Baseline

The baseline lets an existing repo adopt ci-parity one workflow at a time. It records,
per workflow, how many inline-tool (P001) findings are tolerated; counts can only go down.

```sh
ci-parity baseline            # write .ci-parity-baseline.json from the current findings
ci-parity check               # tolerate up to each count; any rise fails with P004
```

A workflow at 0, or missing from the baseline, is locked: any inline tool fails. When a
count drops, `check` prints a ratchet table; run `ci-parity baseline` again to lock in
the lower count. `baseline` refuses to raise a count unless given `--allow-regression`.
`check --strict` ignores the baseline.

The baseline is a count, not a list: in a workflow with a count of N, the first N
inline-tool findings in file order are tolerated. Adding an inline step above existing
ones moves which findings are tolerated, so the error is reported on an older step; the
P004 count is what's new.

## JSON output

`ci-parity check --format json` prints one document. Its shape only changes with a new
`version`.

```json
{
  "version": 1,
  "summary": { "errors": 1, "warnings": 0, "tolerated": 2, "allowed": 0 },
  "findings": [
    {
      "rule": "P001",
      "name": "inline-tool",
      "severity": "error",
      "file": ".github/workflows/ci.yml",
      "line": 10,
      "col": 14,
      "job": "web",
      "action": null,
      "step": "Test",
      "command": "pnpm test",
      "message": "step \"Test\" in job \"web\" runs 'pnpm test' directly; already available as 'task test'",
      "suggestion": "task test",
      "reason": null
    }
  ],
  "ratchet": {
    ".github/workflows/ci.yml": { "baseline": 2, "current": 3, "locked": false }
  }
}
```

| Field | Type | Meaning |
|-------|------|---------|
| `version` | number | Schema version, currently `1`. |
| `summary` | object | Finding counts by severity. |
| `findings[].rule`, `.name` | string | Rule ID (`P001`..`P005`) and name. |
| `findings[].severity` | string | `error`, `warning`, `tolerated` or `allowed`. JSON includes all four. |
| `findings[].file`, `.line`, `.col` | string, number | Repo-relative path with forward slashes; 1-based position. |
| `findings[].job` | string or null | Job ID; null for composite actions and P004. |
| `findings[].action` | string or null | Action path for findings in a composite action. |
| `findings[].step` | string or null | The step's `name:`; null when the step has none. |
| `findings[].command` | string or null | The normalized inline command, for P001. |
| `findings[].message` | string | The same text the other formats print. |
| `findings[].suggestion` | string or null | A `task <name>` to use instead, for P001 and P002. |
| `findings[].reason` | string or null | The allow reason, for allowed findings. |
| `ratchet` | object or null | Per scanned file: baseline count, current count, and whether it is locked at 0. Null when no baseline applies (no file, or `--strict`). |

Findings are sorted by file, line, column and rule ID.

## SARIF output

`ci-parity check --format sarif` prints a [SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/)
log, so the findings can appear in GitHub code scanning (the Security tab) alongside the
annotations. Errors are `error`, warnings `warning` and tolerated findings `note`; an
allowed finding is included with its reason as an in-source suppression, so code scanning
shows it as dismissed. Each result carries a fingerprint built from the rule, file, job,
step and command, so moving lines doesn't open a new alert.

With the action:

```yaml
jobs:
  ci-parity:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
    steps:
      - uses: actions/checkout@v7
      - uses: jimmykurian/ci-parity@v0
        id: parity
        with:
          format: sarif
          fail-on-error: false
      - uses: github/codeql-action/upload-sarif@v4
        with:
          sarif_file: ${{ steps.parity.outputs.sarif-file }}
          category: ci-parity
```

Code scanning is free for public repositories; private ones need GitHub Advanced
Security. With `fail-on-error: false` the job succeeds and code scanning reports the
findings; leave it `true` to also fail the job.

## FAQ

**Why not actionlint or zizmor?** Use them too. actionlint checks workflow syntax and
runs shellcheck on scripts, and zizmor finds security problems; neither knows about your
task runner. ci-parity checks that CI and local runs use the same commands. This
repository runs all three.

**What about task2ci?** It generated workflows from Taskfile annotations, and it's
archived. ci-parity keeps hand-written workflows and checks them instead, with a baseline
so an existing repository can adopt it gradually.

**What about Make, just or mise?** Make and just are supported with `runner: make` and
`runner: just`; see [Other runners](#other-runners). A mise reader is in progress. In a repository
using one runner, the others count as inline tools, because they bypass it.

**How do I allow a step?** With a `# ci-parity: allow <reason>` comment, or an `allow`
entry in `.ci-parity.yml`; see [Allowing a step](#allowing-a-step). To stop flagging a tool
everywhere, add it to `tools.remove`.

**Are PowerShell or cmd steps checked?** No. They're reported as P005 warnings and
skipped, as are steps in any shell other than bash and sh.

## Development

Every command runs through the [Taskfile](Taskfile.yml); CI calls the same tasks.
`task ci` runs formatting, lint, tests, build, self-check and scrub. The linters and
formatters, from golangci-lint to typos, are pinned in [mise.toml](mise.toml) at the
versions CI uses; with [mise](https://mise.jdx.dev/), `mise install` sets them up.
`task hooks:install` turns on the pre-commit, commit-msg and pre-push hooks in
`.githooks/`. The devcontainer (`.devcontainer/`) sets all of this up in VS Code or
Codespaces.
See [CONTRIBUTING.md](CONTRIBUTING.md) for how to propose and test a change.

## License

[Apache-2.0](LICENSE)
