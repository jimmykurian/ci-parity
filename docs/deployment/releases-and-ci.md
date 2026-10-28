<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Releases and CI

What runs on pull requests and on `main`, how a release is cut, and how the GitHub Action
and its binaries are published.

## Continuous integration

`.github/workflows/ci.yml` runs on pull requests and on pushes to `main`:

| Job | Runs |
| --- | --- |
| **Checks** | checkout, setup-go (from `go.mod`), mise-action (tools from `mise.toml`), gitleaks-action, then one `task` step per check in `task ci` except the self-check. Every step runs even after one fails. |
| **Self-check** | `task build`, then the action from this checkout (`uses: ./` with `binary: ./bin/ci-parity`) on this repository, and on the `inline-tools` fixture to check its outputs (`task action:expect`). On `main`, also a SARIF run uploaded to code scanning. |
| **Check table** | On pull requests, one comment listing every check's result, updated on each push. |
| **Badges** | On `main`, one shields.io endpoint file per check, committed to the `badges` branch for the README's status table. |

`.github/workflows/labeler.yml` labels pull requests by the paths they change
(`.github/labeler.yml`), and `.github/workflows/pr-assign-author.yml` assigns each new one
to its author, or a bot's to the owner.

The "Protect main" ruleset requires squash-merged pull requests, signed commits, linear
history, and the **Checks** and **Self-check** jobs; it blocks force pushes and deletion,
and repository admins can bypass it.

Every `run:` step calls `task`, and every third-party action is pinned to a full commit
SHA with a version comment. The repository checks its own workflows with ci-parity:
`.ci-parity.yml` and a `.ci-parity-baseline.json` with every workflow locked at 0.

## Tools

The linters and formatters are pinned in `mise.toml`, never in `go.mod`. CI installs them
with a SHA-pinned `jdx/mise-action`; locally, `mise install` does. Each `task lint:<group>`
checks its tools are on `PATH` and names the one missing.

## Cutting a release

`.github/workflows/release.yml` runs on pushes to `main`:

1. Release Please opens or updates one release pull request, `chore: release X.Y.Z`,
   bumping `.github/.release-please-manifest.json` and `CHANGELOG.md` from the
   Conventional Commits since the last release. Before 1.0, a `feat` is a minor release
   and a `fix` a patch; `docs` and `refactor` are listed, `chore`, `ci`, `build`, `test` and
   `style` are not.
2. It runs with `GITHUB_TOKEN`, so the release pull request starts no checks on its own:
   close and reopen it, wait for the checks, then merge it.
3. Merging tags `vX.Y.Z` and publishes the GitHub Release. In the same workflow, goreleaser
   (`.goreleaser.yaml`) builds linux, darwin and windows on amd64 and arm64, with the
   version, commit and date stamped in, and attaches
   `ci-parity_<version>_<os>_<arch>.tar.gz` (zip on windows) and `checksums.txt`. The
   build uses no caches.
4. `task release:float` moves the `v0` tag to the release, so
   `uses: jimmykurian/ci-parity@v0` follows it. Pre-releases don't move it.
5. To update the Marketplace listing, edit the new release on GitHub, keep "Publish this
   Action to the GitHub Marketplace" ticked, and save.

`task release:snapshot` builds every archive into `dist/` locally, and
`task lint:release` checks `.goreleaser.yaml`.

## See also

- [Documentation index](../README.md)
- [Flows](../architecture/flows.md#from-pull-request-to-release)
- [GitHub Action](../reference/action.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)

Last reviewed: 2026-10-28
