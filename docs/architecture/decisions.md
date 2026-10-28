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
- **SARIF keeps allowed findings as suppressions** and gives each finding a fingerprint
  from its rule, file, job, step and command, so code scanning dismisses rather than
  forgets them and moving lines doesn't open a new alert.

## How it changes files

- **`fix` only rewrites what can't change meaning:** a one-line value whose command, after
  any wrappers and `NAME=value` prefixes, is exactly the flagged command. Wrappers,
  quotes and comments stay; each edit is re-checked against the file before writing.

## How it is built

- **Clean architecture with ports.** Rules and use cases never import a parser or a file
  format, so a new runner or output format is a new adapter. See [Packages](packages.md).
- **Four Go dependencies:** `gopkg.in/yaml.v3`, `mvdan.cc/sh/v3`,
  `github.com/BurntSushi/toml` (mise config) and `github.com/google/go-cmp` (tests). Dev
  tools are pinned in `mise.toml`, never in `go.mod`.
- **Golden fixtures drive the tests.** Each case is a miniature repository with its
  expected output in every format, so behavior is reviewed as text diffs.

## How it is released

- **Release Please with `GITHUB_TOKEN`.** No app or token to manage; the cost is that the
  release pull request starts no checks until it is closed and reopened.
- **goreleaser in the same workflow,** because a tag created with `GITHUB_TOKEN` starts no
  other workflow. The release job uses no caches, and every action is pinned to a commit
  SHA.
- **`v0` floats** to the newest release, so `uses: jimmykurian/ci-parity@v0` picks up
  fixes; pin a version or SHA to avoid that.

## Qualities

- A repository with 50 workflows checks in well under a second.
- Output is identical across platforms: paths use forward slashes and findings are
  sorted by file, line, column and rule.
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

Last reviewed: 2026-10-28
