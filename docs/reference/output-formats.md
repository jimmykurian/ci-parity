<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Output formats

The text, `github`, JSON and SARIF formats of `ci-parity check`, exactly as they are printed.

## Formats

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

## See also

- [Reference index](README.md)
- [Rules](rules.md)
- [GitHub Action](action.md)

Last reviewed: 2026-10-28
