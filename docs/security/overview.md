<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Security overview

How ci-parity treats the repositories it reads, and how the project protects its code,
releases and CI. To report a vulnerability, follow [SECURITY.md](../../SECURITY.md).

## Threat model

ci-parity runs in CI on files that may come from an untrusted pull request. Its inputs are
workflows, composite actions, runner files, `.ci-parity.yml` and the baseline.

- **No execution.** It never runs the commands it reads, the task runner, or a shell; it
  parses them. The code has no `os/exec` or network imports outside tests.
- **No network at check time.** Results depend only on the files and `-C`.
- **Escaped output.** The `github` format escapes workflow-command characters, so a crafted
  step name or message can't inject a workflow command; JSON and SARIF are encoded as
  data.
- **Bounded reading.** Includes, imports and modules are followed to depth 10, and cycles
  are cut.
- **Writes only on `fix`,** only to discovered workflow files inside the repository, and
  only where the expected text is still in place.

## The action

- Every input reaches the scripts through `env:`, never pasted into them, so an input can't
  inject shell code.
- The binary comes from the pinned release with `gh release download` and is checked
  against `checksums.txt` before it runs.

## Supply chain

- Four Go dependencies, each recorded with its reason; `govulncheck` checks them and the
  standard library in CI.
- Every third-party action is pinned to a full commit SHA; zizmor and actionlint check the
  workflows.
- Dev tools are pinned in `mise.toml`; mise's npm provenance check is not bypassed.
- The release job uses no caches, so a cache written elsewhere can't reach the binaries.

## Repository controls

- gitleaks scans the history on every pull request and push, and runs in the pre-commit
  hook. A hit is fixed, never allowlisted.
- `main` takes only squash-merged pull requests with passing checks and signed commits.
- Workflows start with read-only or no permissions; each job that needs more asks for it.

## See also

- [Documentation index](../README.md)
- [SECURITY.md](../../SECURITY.md)
- [Design decisions](../architecture/decisions.md)

Last reviewed: 2026-10-28
