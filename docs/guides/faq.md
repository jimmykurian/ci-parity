<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# FAQ

Short answers to common questions, with links to the details.

## Why not actionlint or zizmor?

Use them too. actionlint checks workflow syntax and runs shellcheck on scripts, and zizmor
finds security problems; neither knows about your task runner. ci-parity checks that CI
and local runs use the same commands. See [Why ci-parity](why.md).

## What about task2ci?

It generated workflows from Taskfile annotations, and it's archived. ci-parity keeps
hand-written workflows and checks them instead, with a baseline so an existing repository
can adopt it gradually.

## Does it support Make, just or mise?

Yes: `runner: make`, `runner: just` or `runner: mise`; see [Configuring](configuring.md#runners)
and the [recipes](recipes.md). In a repository using one runner, the others count as
inline tools, because they bypass it.

## Does it run my task runner or my scripts?

No. It reads workflows and runner files as text, never runs `task`, `make`, `just`,
`mise` or anything in a `run:` step, and makes no network calls while checking.

## How do I allow a step?

With a `# ci-parity: allow <reason>` comment on the step's line or the line above, or an
`allow` entry in `.ci-parity.yml`; see [Allowing a step](configuring.md#allowing-a-step).
To stop flagging a tool everywhere, add it to `tools.remove`.

Version checks, tool installs pinned to a version, sign-in commands and inline scripts
are already allowed by the
[built-in allow set](../reference/configuration.md#built-in-allow-set); set
`builtin_allow: false` to have them reported.

## Why isn't my command flagged?

Only tools on the list are flagged, matched on the command's basename; anything else is
ignored. Add yours with `tools.extend`; see [Configuring](configuring.md#tools).

## Are PowerShell or cmd steps checked?

No. They're reported as [P005](../rules/p005.md) warnings and skipped, as are steps in
any shell other than bash and sh. A job on a Windows runner defaults to `pwsh`, so its
steps are skipped unless they set `shell: bash`.

## What about steps in composite actions?

Local composite actions under `.github/actions/` are checked like workflow jobs, and
their findings name the action instead of a job.

## The error is on a step I didn't change. Why?

With a baseline, the first N inline-tool findings in a workflow are tolerated. A new
inline step above older ones moves which finding is over the count. See
[Adopting with a baseline](baseline.md#things-to-know).

## See also

- [Guides](README.md)
- [Rules](../rules/README.md)

Last reviewed: 2026-10-29
