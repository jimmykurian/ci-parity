<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Reference

What ci-parity reads, what it reports, and the exact formats it prints. These pages are
the contract: rule IDs, messages, config keys, the baseline format and the JSON and SARIF
schemas change only with a recorded decision and a new version.

| Page | Covers |
| --- | --- |
| [Configuration](configuration.md) | The files scanned, and every `.ci-parity.yml` key with its default |
| [Baseline](baseline.md) | `.ci-parity-baseline.json` and how counts tolerate, lock and ratchet findings |
| [Analysis](analysis.md) | Steps, shells, command extraction, wrappers, allow comments and invokers |
| [Runners](runners.md) | Reading Taskfiles, Makefiles, justfiles and mise tasks, and each runner's call syntax |
| [Rules](rules.md) | P001 to P005, their exact messages, and how suggestions are chosen |
| [CLI](cli.md) | `check`, `baseline`, `fix` and `version`, their flags, and the exit codes |
| [Output formats](output-formats.md) | The text, `github`, JSON and SARIF formats |
| [GitHub Action](action.md) | `action.yml`'s inputs, outputs and steps |
| [Compatibility](compatibility.md) | What stays stable between releases, what may change, deprecation, and the platforms tested |

## Vocabulary

- **Runner:** the task runner: Taskfile, make, just or mise.
- **Invoker:** how CI calls the runner. A shell invoker (`task x`, `make x`, `just x`,
  `mise run x`) or an action invoker (a local composite action such as
  `./.github/actions/run-task` whose input names the task).
- **Tool:** a command on the tools list. Only tools are flagged; everything else is
  ignored.
- **Finding:** one rule hit at one location, with severity `error`, `warning`,
  `tolerated` or `allowed`.
- **Baseline:** the per-workflow count of tolerated P001 findings. Counts only go down. A
  count of zero, or a workflow missing from the baseline, locks that workflow.

## See also

- [Documentation index](../README.md)
- [Design decisions](../architecture/decisions.md)

Last reviewed: 2026-10-28
