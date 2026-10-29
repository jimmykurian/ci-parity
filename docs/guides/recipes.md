<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Recipes

A minimal setup for each task runner that ci-parity passes: the runner file, the config
when it isn't Taskfile, and a workflow that calls the runner. Each one checks clean with
`0 errors, 0 warnings, 0 tolerated, 0 allowed`.

## Taskfile

```yaml
# Taskfile.yml
version: "3"

tasks:
  install:
    cmds:
      - pnpm install --frozen-lockfile
  lint:
    cmds:
      - pnpm lint
  test:
    cmds:
      - pnpm test
```

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: pnpm/action-setup@v6
      - uses: go-task/setup-task@v2
      - run: task install
      - run: task lint
      - run: task test
      - uses: jimmykurian/ci-parity@v1
```

Taskfile is the default runner, so no config is needed. Setup, such as installing pnpm
and task, goes in `uses:` steps, which ci-parity doesn't flag.

## make

```yaml
# .ci-parity.yml
runner: make
```

```makefile
# Makefile
.PHONY: lint test

lint:
	golangci-lint run

test:
	go test ./...
```

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - run: make lint
      - run: make test
      - uses: jimmykurian/ci-parity@v1
```

## just

```yaml
# .ci-parity.yml
runner: just
```

```just
# justfile
lint:
    cargo clippy -- -D warnings

test:
    cargo test

[private]
release-notes:
    git cliff --latest
```

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: extractions/setup-just@v4
      - run: just lint
      - run: just test
      - uses: jimmykurian/ci-parity@v1
```

`release-notes` is `[private]`, so calling it from CI would be a
[P003](../rules/p003.md).

## mise

```yaml
# .ci-parity.yml
runner: mise
```

```toml
# mise.toml
[tools]
python = "3.13"
uv = "0.9"

[tasks.lint]
run = "uv run ruff check ."

[tasks.test]
run = "uv run pytest"
```

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: jdx/mise-action@v5
      - run: mise run lint
      - run: mise run test
      - uses: jimmykurian/ci-parity@v1
```

Package runners are looked through when suggesting a task, so an inline `run: pytest`
here is reported with the task that runs `uv run pytest`:

```text
.github/workflows/ci.yml
  11:14   P001 inline-tool   step 4 in job "test" runs 'pytest' directly; already available as 'mise run test'
```

## See also

- [Guides](README.md)
- [Configuring](configuring.md#runners)
- [Runners reference](../reference/runners.md)

Last reviewed: 2026-10-29
