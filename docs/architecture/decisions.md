<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Design decisions

The decisions that shape ci-parity and the reason for each. Changing one of them, or the
contract in the [reference](../reference/README.md), starts with a proposal in an issue
or pull request and the maintainer's agreement before any code.

## How it reads a repository

- **Statically, with no network and no execution.** ci-parity parses the runner's files
  itself and never runs `task`, `make`, `just` or `mise`, so results don't depend on what
  is installed and a check can't run untrusted code. Output depends only on the files and
  `-C`.
- **Tools are matched by basename, through transparent wrappers.** `./gradlew` and
  `sudo -E env CI=1 pnpm test` are both seen for what they run. Each wrapper's
  value-taking flags are listed so their values aren't mistaken for the command.
- **Only bash and sh are parsed.** Other shells are reported as P005 rather than guessed
  at.
- **Positions are exact.** Lines and columns map back through folded YAML and `${{ }}`
  expressions, so annotations land on the right character.

## What it reports

- **The baseline is a count per workflow, not a list.** Counts survive reformatting and
  only go down. The cost: the *first* N findings in a workflow are tolerated, so an
  inline step added above older ones moves which finding is reported.
- **Messages are a contract.** People search for them, so their wording is fixed, and
  each runner's own words are used: `make lint`, `'release' is private`.
- **Suggestions look through package runners.** `npx eslint .` matches a task that runs
  `eslint .`, on either side, because that is how most repositories call their tools.
- **Suggestions name only tasks that do exactly what the step does.** A task with one
  command, no dependencies and no calls to other tasks; otherwise `fix` would turn a
  format check into the whole CI run.
- **Setup and info commands are allowed by default, not dropped.** Versions, pinned
  installs, sign-in and inline scripts were a quarter of the field trial's errors, and
  none belongs in a task. They are reported as `allowed` with a built-in reason, so they
  stay visible, and `builtin_allow: false` restores them as errors.
- **SARIF keeps allowed findings as suppressions** and gives each finding a fingerprint
  from its rule, file, job, step and command, so code scanning dismisses rather than
  forgets them and moving lines doesn't open a new alert.

## How it changes files

- **`fix` only rewrites what can't change meaning:** a one-line value whose command, after
  any wrappers and `NAME=value` prefixes, is exactly the flagged command. Wrappers,
  quotes and comments stay; each edit is re-checked against the file before writing.

## How it is built

- **`go.mod`'s `go` line is the lowest Go that builds it, 1.26 (mvdan/sh's floor), and
  its `toolchain` line is the release CI builds with.** `go install …@latest` reads only
  the `go` line, so Go 1.26 users don't download a newer toolchain.
- **Every file carries an SPDX license header,** `SPDX-License-Identifier: Apache-2.0`
  under a copyright line, so a file copied out of the repository keeps its license.
  "All rights reserved" is left out: it contradicts the license. `task lint:headers`
  checks it.
- **Clean architecture with ports.** Rules and use cases never import a parser or a file
  format, so a new runner or output format is a new adapter. See [Packages](packages.md).
- **Three Go dependencies in the binary,** `gopkg.in/yaml.v3`, `mvdan.cc/sh/v3` and
  `github.com/BurntSushi/toml` (mise config), and three for tests:
  `github.com/google/go-cmp` for golden diffs, and `github.com/onsi/ginkgo/v2` and
  `github.com/onsi/gomega` for the integration and end-to-end specs. Dev tools are pinned
  in `mise.toml`, never in `go.mod`.
- **BDD specs for the outer layers only.** The integration and end-to-end tests are
  Ginkgo specs: they describe behavior the way the reference does, and `gexec` builds and
  runs the binary. Unit tests stay `testing` table tests, which is what Go contributors
  expect, and they need to sit in their package to reach unexported functions.
- **Golden fixtures drive the tests.** Each case is a miniature repository with its
  expected output in every format, so behavior is reviewed as text diffs.

## How it is released

- **Release Please with `GITHUB_TOKEN`.** No app or token to manage; the cost is that the
  release pull request starts no checks until it is closed and reopened.
- **goreleaser in the same workflow,** because a tag created with `GITHUB_TOKEN` starts no
  other workflow. The release job uses no caches, and every action is pinned to a commit
  SHA.
- **The major tag floats:** `v1` follows the newest 1.x release, so
  `uses: jimmykurian/ci-parity@v1` picks up fixes; pin a version or SHA to avoid that.
- **A Starlight site built from `docs/`, on GitHub Pages.** `docs/` stays the single
  source and keeps rendering on GitHub; a sync step turns each H1 into a page title and
  rewrites links, so no page is written twice. Starlight (Astro) brings search, light and
  dark themes and accessible navigation with fewer dependencies than Docusaurus. The
  site's npm packages are for the site alone: nothing from them reaches the binary or
  `go.mod`.
- **Install routes that need nothing new per release:** the Action, the release
  binaries, `go install`, mise's GitHub backend, and a pre-commit hook
  (`.pre-commit-hooks.yaml`) that builds from the tag. A Homebrew tap waits for demand:
  it needs a second repository and a token on the release job.

## Qualities

- A repository with 50 workflows checks in well under a second: `test/perf` times a
  generated one (1,200 steps) on every OS in CI and fails above one second, and
  `task bench` reports the numbers.
- Output is identical across platforms: paths use forward slashes and findings are
  sorted by file, line, column and rule. CI runs every test on Linux, macOS and Windows.
- **The baseline stays a count.** Fingerprint lists (one entry per tolerated finding)
  would name exactly the new step, but they'd mean a breaking v2 format whose entries
  change whenever a step is reformatted. See [Baseline: why a count](../reference/baseline.md#why-a-count).
- **Code metrics are checked, not proposed.** The README's metrics (scc lines of code, Go
  function complexity from `tools/metrics`) fail CI when stale, so each pull request
  updates them. A bot pull request would need a GitHub App: one opened with
  `GITHUB_TOKEN` starts no checks.
- **The formats have JSON Schemas** in `schemas/`, published with the docs site, checked
  against every fixture by `task lint:schemas`, and the config schema's keys are tested
  against the parser's.
- A config error names the key at fault and exits 2.

## Ideas not yet planned

- Reading `task --list-all --json` when task is installed, for Taskfile features that
  can't be resolved statically, such as templated include paths.
- A companion `run-task` action, so workflows can call tasks through a `uses:` step
  without writing their own.

## See also

- [Architecture](README.md)
- [Reference](../reference/README.md)
- [Security overview](../security/overview.md)

Last reviewed: 2026-10-29
