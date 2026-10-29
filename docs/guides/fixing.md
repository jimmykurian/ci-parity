<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Fixing findings

Let `ci-parity fix` rewrite inline steps to the task that already runs the same command,
and fix the rest by hand.

## What it changes

On this workflow, next to a Taskfile with `install`, `lint`, `test` and `build` tasks:

```yaml
    steps:
      - uses: actions/checkout@v7
      - name: Install
        run: pnpm install --frozen-lockfile
      - name: Lint
        run: sudo pnpm lint
      - name: Test
        run: "pnpm test"
      - name: Build and test
        run: pnpm build && pnpm test
```

`--dry-run` lists the edits without writing them; without it, `fix` makes them:

```text
$ ci-parity fix
.github/workflows/ci.yml:7:14 'pnpm install --frozen-lockfile' -> 'task install'
.github/workflows/ci.yml:9:19 'pnpm lint' -> 'task lint'
.github/workflows/ci.yml:11:15 'pnpm test' -> 'task test'
fixed 3 steps in 1 file; 2 inline-tool findings need a manual change
```

```yaml
      - name: Install
        run: task install
      - name: Lint
        run: sudo task lint
      - name: Test
        run: "task test"
      - name: Build and test
        run: pnpm build && pnpm test
```

The wrapper and the quotes stay, and so does everything else in the file. The last step
is left for you, and `check` still reports it with the task to use:

```text
$ ci-parity check
.github/workflows/ci.yml
  13:14   P001 inline-tool   step "Build and test" in job "web" runs 'pnpm build' directly; already available as 'task build'
  13:28   P001 inline-tool   step "Build and test" in job "web" runs 'pnpm test' directly; already available as 'task test'

2 errors, 0 warnings, 0 tolerated, 0 allowed
```

## What it leaves alone

`fix` only changes a step when that can't change what the step does:

- The `run:` value is on one line: plain, quoted (`"pnpm test"`), or a `run: |` block
  with one line.
- Before the command there are only wrappers such as `sudo`, `env` and `timeout` with
  their options, or variables like `CI=1`. Those stay, so `sudo pnpm lint` becomes
  `sudo task lint`.
- The command matches a task exactly, which is when [P001](../rules/p001.md) shows
  "already available as". That task runs in the step's `working-directory`, and calling
  the runner from there finds it; otherwise there's no suggestion and nothing to fix.

Steps with `&&`, pipes, redirects, several lines or `${{ }}` are counted as needing a
manual change. Each edit is checked against the file again just before it's written.

## With a baseline

Fixing lowers the counts, so run `ci-parity baseline` afterwards to lock them in; `fix`
reminds you when that applies. See [Adopting with a baseline](baseline.md).

## See also

- [Guides](README.md)
- [P001 inline-tool](../rules/p001.md)
- [CLI reference](../reference/cli.md)

Last reviewed: 2026-10-28
