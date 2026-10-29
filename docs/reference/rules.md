<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Rules

Rules P001 to P005: when each fires, its severity, its exact message, and how suggestions are chosen.

Each rule also has a page with an example and how to fix or allow it:
[P001](../rules/p001.md), [P002](../rules/p002.md), [P003](../rules/p003.md),
[P004](../rules/p004.md) and [P005](../rules/p005.md).

## The rules

| ID   | Name             | Severity                         | Trigger |
|------|------------------|----------------------------------|---------|
| P001 | inline-tool      | error, or tolerated via baseline, or allowed | A `run:` step calls a listed tool directly. |
| P002 | unknown-task     | error                            | An invoker names a task the runner does not define. |
| P003 | internal-task    | error                            | An invoker names an internal task: `internal: true` (Taskfile), private (just) or hidden (mise). |
| P004 | regression       | error                            | A workflow's P001 count exceeds its baseline, or a locked workflow has any P001. One finding per workflow, located at line 1 of the workflow. |
| P005 | unsupported      | warning                          | Unsupported shell, shell parse error, dynamic command name, allow comment without reason. |

Message wording (normative, because users grep for it). With another runner ([runners](runners.md#make-just-and-mise)), its
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

## Suggestions

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

- P001: the task must also run in the step's working directory
  ([analysis](analysis.md#working-directory)), and calling the runner from that directory
  must read the same root file, so the suggested name exists there. Where each runner's
  tasks run is in [runners](runners.md#where-tasks-run). When either is unknown, nothing is
  suggested, so `fix` leaves the step alone.
- P001: the task must run that one command and nothing else: no other command, no
  dependency, and no call to another task. A task that does more would change what the
  step runs, so it is never suggested.

  | Runner | Does more than its command |
  | --- | --- |
  | Taskfile | `deps:`, or a `task:`, `for:` or `defer:` entry in `cmds:` |
  | just | dependencies, before or after `&&` |
  | make | prerequisites, order-only ones included, on any of the target's rules |
  | mise | `depends`, `depends_post`, or a `run` entry that isn't a string |

- P002: a task name that has the unknown name as a prefix wins, shortest first.
  Otherwise suggest the name with the lowest Levenshtein distance, when that distance is
  at most 3 and less than the length of the unknown name. Internal tasks are never
  suggested.

## See also

- [Reference index](README.md)
- [Analysis](analysis.md)
- [Output formats](output-formats.md)

Last reviewed: 2026-10-29
