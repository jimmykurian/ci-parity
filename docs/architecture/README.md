<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Architecture

How ci-parity is put together: its context, its packages and ports, the main flows, and
the decisions behind them. Diagrams are Mermaid, which GitHub renders in place.

| Page | Shows |
| --- | --- |
| [System context](#system-context) | Who runs ci-parity, what it reads and where its results go |
| [Packages](packages.md) | The layers, their ports, and which adapter implements each |
| [Flows](flows.md) | A `check`, a `fix`, and a change from pull request to release |
| [Design decisions](decisions.md) | Why it works the way it does, and ideas not yet planned |

## System context

```mermaid
flowchart LR
    dev["Developer<br/>(local CLI)"] --> cli
    ci["GitHub Actions job<br/>(uses: jimmykurian/ci-parity@v1)"] --> action
    action["Composite action<br/>action.yml"] -->|"downloads and verifies"| rel["GitHub Release<br/>binaries and checksums.txt"]
    action --> cli["ci-parity binary"]
    cli -->|reads| repo["Repository files<br/>workflows, composite actions,<br/>Taskfile / Makefile / justfile / mise,<br/>.ci-parity.yml, baseline"]
    cli -->|"text, json"| out["Terminal or a file"]
    cli -->|"github format"| ann["Annotations on the<br/>changed lines"]
    cli -->|"sarif + upload-sarif"| cs["GitHub code scanning"]
    cli -->|"fix"| repo
```

ci-parity reads only the repository's files. It never runs the task runner or the tools
it finds, and makes no network calls while checking; the action downloads the binary
before the check starts.

## See also

- [Documentation index](../README.md)
- [Reference](../reference/README.md)

Last reviewed: 2026-10-29
