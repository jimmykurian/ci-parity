<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Rules

Each rule ci-parity reports: what it catches, an example, and how to fix or allow it.
Code scanning alerts link here.

| ID | Name | Severity | Reported when |
| --- | --- | --- | --- |
| [P001](p001.md) | inline-tool | error, tolerated or allowed | A `run:` step calls a listed tool directly |
| [P002](p002.md) | unknown-task | error | CI calls a task the runner doesn't define |
| [P003](p003.md) | internal-task | error | CI calls a task the runner marks internal |
| [P004](p004.md) | regression | error | A workflow has more inline tools than its baseline allows |
| [P005](p005.md) | unsupported | warning | A step ci-parity can't analyze |

The exact message wording, and how suggestions are chosen, are in the
[rules reference](../reference/rules.md). Messages are stable, so you can search for
them.

## See also

- [Rules reference](../reference/rules.md)
- [Guides](../guides/README.md)

Last reviewed: 2026-10-28
