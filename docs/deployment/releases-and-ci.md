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
| **Code metrics** | A step in **Checks**: `task metrics:check` fails when the README's metrics table doesn't match the code, so each change updates it with `task metrics:readme`. |
| **Tests (macOS, Windows)** | The unit, integration, end-to-end and performance tests on `macos-latest` and `windows-latest`. |
| **Action download** | On Linux, macOS and Windows, the published action, pinned by SHA with no `binary` or `version` input: it downloads the latest release, verifies it against `checksums.txt` and runs it on the `inline-tools` fixture and on this repository. Move the pin to each new release. |
| **Check table** | On pull requests, one comment listing every check's result, updated on each push. |
| **Badges** | On `main`, one shields.io endpoint file per check, committed to the `badges` branch for the README's status table. |

`.github/workflows/labeler.yml` labels pull requests by the paths they change
(`.github/labeler.yml`), and `.github/workflows/pr-assign-author.yml` assigns each new one
to its author, or a bot's to the owner.

The "Protect main" ruleset requires squash-merged pull requests, signed commits, linear
history, and the **Checks**, **Self-check**, **Tests (macos-latest)** and
**Tests (windows-latest)** jobs; it blocks force pushes and deletion, and repository admins
can bypass it. **Action download** isn't required: it tests the published release, which a
pull request can't change, and it can fail for reasons outside the repository, such as a
GitHub outage.

Every `run:` step calls `task`, and every third-party action is pinned to a full commit
SHA with a version comment. The repository checks its own workflows with ci-parity:
`.ci-parity.yml` and a `.ci-parity-baseline.json` with every workflow locked at 0.

## Tools

The linters and formatters are pinned in `mise.toml`, never in `go.mod`. CI installs them
with a SHA-pinned `jdx/mise-action`; locally, `mise install` does. Each `task lint:<group>`
checks its tools are on `PATH` and names the one missing.

## Docs site

`.github/workflows/site.yml` publishes `docs/` to GitHub Pages at
<https://ci-parity.is-a.dev/> on every push to `main` that changes `docs/`,
`site/` or the workflow, and on demand.

The domain is a subdomain of [is-a.dev](https://is-a.dev): `domains/ci-parity.json` in
[is-a-dev/register](https://github.com/is-a-dev/register) points it at
`jimmykurian.github.io` with a CNAME record, and the repository's Pages settings name it as
the custom domain with HTTPS enforced. GitHub redirects the old
`jimmykurian.github.io/ci-parity/` URLs to it.

- **Build:** `task site:install` installs the site's packages from `site/pnpm-lock.yaml`;
  `task site:build` copies `docs/` into the Starlight site with `site/scripts/sync-docs.mjs`
  and builds it, failing on a broken link or anchor. The build uses no caches.
- **Deploy:** `actions/deploy-pages` publishes the built site to the `github-pages`
  environment; only that job can write to Pages.
- **Pull requests:** the **Checks** job runs `task site:test` and `task site:build`, so a
  broken link fails before it reaches `main`.

Preview locally with `task site:dev`, or `task site:build` then `task site:preview`.

## Cutting a release

`.github/workflows/release.yml` runs on pushes to `main`:

1. Release Please opens or updates one release pull request, `chore: release X.Y.Z`,
   bumping `.github/.release-please-manifest.json`, `CHANGELOG.md` and every line of
   `docs/guides/getting-started.md` marked `x-release-please-version` from the
   Conventional Commits since the last release. A breaking change (`!` or a `BREAKING CHANGE` footer) is a major release, a `feat` a minor one, and anything else listed (`fix`,
   `perf`, `docs`, `refactor`) a patch; `chore`, `ci`, `build`, `test` and `style` are
   hidden and cut no release.
2. It runs with `GITHUB_TOKEN`, so the release pull request starts no checks on its own:
   close and reopen it, wait for the checks, then merge it.
3. Merging tags `vX.Y.Z` and creates the GitHub Release as a draft (`draft` and
   `force-tag-creation` in the Release Please config, so the tag exists at once). In the
   same workflow, goreleaser (`.goreleaser.yaml`) builds linux, darwin and windows on
   amd64 and arm64, with the version, commit and date stamped in, attaches
   `ci-parity_<version>_<os>_<arch>.tar.gz` (zip on windows) and `checksums.txt` to the
   draft (`use_existing_draft`), then publishes it. The latest release therefore always
   has its archives, which the action downloads; if the build fails, the release stays a
   draft that users don't see. The build uses no caches.
4. `task release:float` moves the release's major tag, such as `v1`, to it, so
   `uses: jimmykurian/ci-parity@v1` follows it. Pre-releases don't move it, and `v0` stays
   on the last 0.x release.
5. To update the Marketplace listing, edit the new release on GitHub, keep "Publish this
   Action to the GitHub Marketplace" ticked, and save.

`task release:snapshot` builds every archive into `dist/` locally, and
`task lint:release` checks `.goreleaser.yaml`.

## See also

- [Documentation index](../README.md)
- [Flows](../architecture/flows.md#from-pull-request-to-release)
- [GitHub Action](../reference/action.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)

Last reviewed: 2026-10-29
