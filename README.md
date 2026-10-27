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

## FAQ

## Development

Every command runs through the [Taskfile](Taskfile.yml); CI calls the same tasks.
`task ci` runs formatting, lint, tests, build, self-check and scrub. Linting needs
[golangci-lint](https://golangci-lint.run/) v2 and
[actionlint](https://github.com/rhysd/actionlint) on `PATH`.

## License

[Apache-2.0](LICENSE)
