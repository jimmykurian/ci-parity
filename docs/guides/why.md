<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Why ci-parity

What problem it solves, the ways to solve it, and how it relates to the workflow linters
you may already run.

## The problem

A repository's task runner says how to build, test and lint it: `task test` locally runs
what CI should run. Workflows drift from it. A step calls `pnpm test` directly, someone
changes the Taskfile's test task to add a flag, and CI keeps running the old command. A
workflow calls `task lint:all` after the task was renamed, and the step fails only when
that workflow next runs. Each drift is small, and together they mean "it passes locally"
stops predicting "it passes in CI".

## Three ways to keep them in line

| Approach | How | Trade-off |
| --- | --- | --- |
| **Convention** | A rule in CONTRIBUTING and code review | Nothing checks it, so it drifts the moment review misses a step |
| **Generation** | Generate workflows from the task runner's definitions | No drift by construction, but workflows become generated files: hand edits are lost, and features the generator doesn't model are out of reach |
| **Enforcement** | Keep hand-written workflows and lint them against the runner | Workflows stay ordinary YAML; the linter has to read both sides accurately |

ci-parity is the enforcement approach. It reads workflows and the runner's files
statically, flags commands that bypass the runner, checks that every task CI calls exists
and is public, and keeps a [baseline](baseline.md) so an existing repository can reach
zero one workflow at a time.

## Next to other workflow linters

Run these too; they check different things, and none of them knows about your task
runner.

| Tool | Checks | Task runner |
| --- | --- | --- |
| **actionlint** | Workflow syntax, expression types, and shellcheck over `run:` scripts | No |
| **zizmor** | Security problems: injection, excessive permissions, unpinned actions | No |
| **gha-workflow-linter** | That `uses:` references resolve to real refs | No |
| **bitwarden/workflow-linter** | An organization's own rules; a rule can't span a workflow and a Taskfile | No |
| **task2ci** (archived June 2026) | Generated workflows from Taskfile annotations, failing CI when the generated file drifted | Generation only; it never linted hand-written workflows, and had no baseline |
| **ci-parity** | Inline tools, unknown and internal tasks, and a baseline that only shrinks | Taskfile, make, just and mise |

This repository runs actionlint, zizmor and ci-parity on its own workflows.

## See also

- [Guides](README.md)
- [Getting started](getting-started.md)
- [Design decisions](../architecture/decisions.md)

Last reviewed: 2026-10-28
