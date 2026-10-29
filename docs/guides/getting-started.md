<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Getting started

Install ci-parity, run it on a repository, and read what it reports.

![A terminal session: ci-parity check reports three inline tools, each already available as a task; ci-parity fix rewrites the three steps; ci-parity check then reports 0 errors.](../assets/ci-parity-demo.svg)

## Install

### GitHub Action

```yaml
- uses: actions/checkout@v7
- uses: jimmykurian/ci-parity@v0
```

`@v0` follows the latest v0.x release. To pin, use a release tag (`@v0.6.0`) or, better,
its full commit SHA with the tag as a comment. The action downloads that release's binary
for the runner, verifies it against `checksums.txt`, and runs `ci-parity check`. Its
inputs and outputs are in the [GitHub Action reference](../reference/action.md), and
[CI setup](ci.md) shows annotations and code scanning.

### Release binaries

Each [release](https://github.com/jimmykurian/ci-parity/releases) has archives for
linux, darwin and windows on amd64 and arm64, and a `checksums.txt`:

```sh
VERSION=0.8.2 # x-release-please-version
gh release download "v$VERSION" --repo jimmykurian/ci-parity \
  --pattern "ci-parity_${VERSION}_linux_amd64.tar.gz" --pattern checksums.txt
sha256sum --ignore-missing -c checksums.txt
tar -xzf "ci-parity_${VERSION}_linux_amd64.tar.gz"
./ci-parity version
```

### go install

```sh
go install github.com/jimmykurian/ci-parity/cmd/ci-parity@latest
```

A `go install` build reports its version as `dev`; the release binaries carry the
version, commit and date.

### mise

[mise](https://mise.jdx.dev/) installs the release archive for your OS and architecture
from GitHub Releases:

```sh
mise use github:jimmykurian/ci-parity@0.8.2 # x-release-please-version
```

Or pin it in `mise.toml` with the repository's other tools:

```toml
[tools]
"github:jimmykurian/ci-parity" = "0.8.2" # x-release-please-version
```

### pre-commit

With [pre-commit](https://pre-commit.com), add the hook to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/jimmykurian/ci-parity
    rev: v0.8.2 # x-release-please-version
    hooks:
      - id: ci-parity
```

pre-commit builds ci-parity from that tag, installing Go if needed, and runs
`ci-parity check` on the whole repository when a commit changes a workflow, an action,
the config, the baseline or the runner's files. If your config reads files from other
paths, add `always_run: true` under `id: ci-parity`.

With [lefthook](https://lefthook.dev) and ci-parity installed (through mise, say), add a
command to `lefthook.yml`:

```yaml
pre-commit:
  commands:
    ci-parity:
      run: ci-parity check
```

## First run

A workflow that calls pnpm directly, next to a Taskfile that already wraps those commands:

```yaml
# .github/workflows/ci.yml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - name: Install
        run: pnpm install --frozen-lockfile
      - name: Test
        run: pnpm test
      - name: Lint
        run: task lint
```

```yaml
# Taskfile.yml
version: "3"

tasks:
  install:
    cmds:
      - pnpm install --frozen-lockfile
  test:
    cmds:
      - pnpm test
```

```text
$ ci-parity check
.github/workflows/ci.yml
  9:14    P001 inline-tool   step "Install" in job "test" runs 'pnpm install --frozen-lockfile' directly; already available as 'task install'
  11:14   P001 inline-tool   step "Test" in job "test" runs 'pnpm test' directly; already available as 'task test'
  13:19   P002 unknown-task  'task lint' is not defined in Taskfile.yml (closest: test)

3 errors, 0 warnings, 0 tolerated, 0 allowed
```

Each line is a position, a [rule](../rules/README.md) and a message. Here there are two
[inline tools](../rules/p001.md) with the task that already runs them, and one
[unknown task](../rules/p002.md). After calling the tasks instead (and installing task in
a `uses:` step), it's clean:

```yaml
    steps:
      - uses: actions/checkout@v7
      - uses: go-task/setup-task@v2
      - name: Install
        run: task install
      - name: Test
        run: task test
```

```text
$ ci-parity check
0 errors, 0 warnings, 0 tolerated, 0 allowed
```

`ci-parity fix` makes changes like these for you; see [Fixing findings](fixing.md).

## Next steps

- A repository with many inline steps can start from a baseline and shrink it:
  [Adopting with a baseline](baseline.md).
- Make, just or mise instead of Taskfile: [Configuring](configuring.md#runners) and the
  [recipes](recipes.md).
- Running it in CI, with annotations or code scanning: [CI setup](ci.md).

## See also

- [Guides](README.md)
- [CLI reference](../reference/cli.md)

Last reviewed: 2026-10-28
