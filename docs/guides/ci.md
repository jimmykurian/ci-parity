<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# CI setup

Run ci-parity in GitHub Actions: annotations on pull requests, code scanning alerts, or
counts for a later step.

## Annotations on pull requests

The action's default format, `github`, turns each finding into an annotation on the
workflow line that caused it. The job fails when there are errors.

```yaml
name: ci-parity
on: [pull_request]

jobs:
  ci-parity:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@v7
      - uses: jimmykurian/ci-parity@v0
```

Each annotation is one line of workflow commands in the job log:

```text
::error file=.github/workflows/ci.yml,line=7,col=14,title=P001 inline-tool::step "Test" in job "web" runs 'pnpm test' directly; already available as 'task test'
```

Paths are relative to `working-directory`, so annotations line up only when it's the
root of the checked-out repository (the default).

## Code scanning

With `format: sarif`, the action writes a [SARIF](../reference/output-formats.md) log
and gives its path as the `sarif-file` output. Upload it, and the findings appear in the
repository's Security tab as code scanning alerts, each linked to its rule's page here.

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

- Errors are `error` alerts, warnings `warning`, and tolerated findings `note`.
- An allowed finding is uploaded as suppressed in source, with its reason, so code
  scanning shows it as dismissed.
- Each alert carries a fingerprint built from the rule, file, job, step and command, so
  moving lines doesn't open a new one.
- With `fail-on-error: false`, the job succeeds and code scanning reports the findings;
  leave it `true` to also fail the job.

Code scanning is free for public repositories; private ones need GitHub Advanced
Security.

## Counts for later steps

With `fail-on-error: false`, the step always succeeds and its outputs carry the counts:

```yaml
      - uses: jimmykurian/ci-parity@v0
        id: parity
        with:
          fail-on-error: false
      - if: steps.parity.outputs.errors != '0'
        run: echo "::warning::${{ steps.parity.outputs.errors }} ci-parity errors"
```

The outputs are `errors`, `warnings`, `tolerated` and `sarif-file`; the
[action reference](../reference/action.md) has every input and output. For your own
tooling, `ci-parity check --format json` prints a document whose shape changes only with
its `version`; see the [output formats](../reference/output-formats.md).

## A baseline in CI

Commit `.ci-parity-baseline.json`, and the action uses it like `check` does: tolerated
findings pass, and a rise fails with [P004](../rules/p004.md). Set `strict: true` to
ignore the baseline, for example in a scheduled job that tracks what's left. See
[Adopting with a baseline](baseline.md).

## See also

- [Guides](README.md)
- [GitHub Action reference](../reference/action.md)
- [Output formats](../reference/output-formats.md)

Last reviewed: 2026-10-28
