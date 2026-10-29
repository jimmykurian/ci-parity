<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Flows

What happens during a `check`, a `fix`, and a change on its way from a pull request to a
release.

## A check

`ci-parity check` builds the adapters for the repository at `-C`, then runs the check use
case (`internal/app/check.go`).

```mermaid
sequenceDiagram
    participant CLI as internal/cli
    participant App as app.Check
    participant Cfg as ConfigSource
    participant WF as WorkflowSource
    participant TL as TaskLoader
    participant Rules as rules.Check
    participant BL as BaselineStore
    participant Rep as Reporter
    CLI->>App: config path, baseline path, strict, reporter
    App->>Cfg: load .ci-parity.yml (or defaults)
    App->>WF: discover and parse workflows and composite actions
    App->>TL: read the runner's files
    App->>Rules: steps, tasks, command extractor
    Rules-->>App: sorted findings
    alt a baseline file exists and not --strict
        App->>BL: read counts
        App->>App: tolerate up to each count, add P004 for rises
    end
    App->>Rep: findings and ratchet rows
    App-->>CLI: errors? exit 1 : exit 0
```

## A fix

`ci-parity fix` runs the same analysis, then rewrites the findings `rules` marked as safe
to change (`internal/app/fix.go`).

```mermaid
sequenceDiagram
    participant CLI as internal/cli
    participant App as app.Fix
    participant Rules as rules.Check
    participant Ed as WorkflowEditor
    CLI->>App: config path, dry run
    App->>Rules: analyze the repository
    Rules-->>App: findings, some carrying an edit
    loop each file with edits
        App->>Ed: apply edits
        Ed->>Ed: check each edit's text is still at its position
        Ed-->>App: edits made
    end
    App-->>CLI: edits, files, findings needing a manual change
```

## From pull request to release

```mermaid
sequenceDiagram
    actor Dev as Contributor
    participant PR as Pull request
    participant CI as ci.yml
    participant Main as main
    participant RP as release.yml
    participant GH as GitHub Releases
    Dev->>PR: push a branch (hooks ran locally)
    PR->>CI: Checks, Self-check, Check table, labels, assignee
    CI-->>PR: one comment with every check's result
    Dev->>Main: squash merge (ruleset: checks pass, signed, linear)
    Main->>CI: Checks, Self-check, Badges, code-scanning upload
    Main->>RP: Release Please opens or updates the release PR
    Dev->>PR: close and reopen the release PR so its checks run, then merge
    RP->>GH: tag vX.Y.Z and create the release as a draft
    RP->>GH: goreleaser attaches archives and checksums.txt, then publishes it
    RP->>Main: move the v0 tag to the release
```

The Marketplace listing follows a release only after its edit page is saved with
"Publish this Action to the GitHub Marketplace" ticked.

## See also

- [Architecture](README.md)
- [Packages](packages.md)
- [Releases and CI](../deployment/releases-and-ci.md)

Last reviewed: 2026-10-28
