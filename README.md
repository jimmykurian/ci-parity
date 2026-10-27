# ci-parity

> **Pre-release.** ci-parity is under active development and has no release yet. The
> commands, config keys and output formats below may change before v0.1.0.

ci-parity checks that GitHub Actions workflows invoke project commands through the
repository's task runner instead of calling build, test, lint and deploy tools inline.
It reports every inline invocation, verifies that every task CI references exists, and
keeps a per-workflow baseline so an existing repo can migrate one workflow at a time
without regressing.

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

## Quick start

## Configuration

## Rules

## Baseline

## FAQ

## Development

Every command runs through the [Taskfile](Taskfile.yml); CI calls the same tasks.
`task ci` runs formatting, lint, tests, build, self-check and scrub. Linting needs
[golangci-lint](https://golangci-lint.run/) v2 and
[actionlint](https://github.com/rhysd/actionlint) on `PATH`.

## License

[Apache-2.0](LICENSE)
