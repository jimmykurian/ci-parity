<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Field trial

ci-parity run on 10 public repositories, with every finding read by hand, to measure how
often it's wrong before 1.0. `task trial` repeats it.

## Running it

```sh
task trial
```

It builds `./bin/ci-parity`, fetches each repository in
[`test/trial/repos.yml`](../../test/trial/repos.yml) at its pinned commit into
`.cache/trial/`, runs `ci-parity check --format json` with that repository's runner, and
prints a Markdown report: findings per repository and rule, then every P002 and every
P001 suggestion. Fetching needs the network; ci-parity only reads the files, and nothing
from the repositories runs. A later run reuses the checkouts.

The commits stay pinned, so a re-run after a fix changes only what ci-parity changed. Add
a repository by adding an entry with its full commit SHA and runner.

## The repositories

| Runner | Repositories | Why |
| --- | --- | --- |
| Taskfile | go-task/task, goreleaser/goreleaser, arduino/arduino-cli | task itself, a large Go project, and task called through `go tool` |
| make | prometheus/prometheus, traefik/traefik, cli/cli | a shared `Makefile.common` with pattern rules, a monorepo, PowerShell steps and a composite action |
| just | casey/just, oxc-project/oxc | just itself, and a large monorepo with composite actions |
| mise | jdx/mise, jdx/hk | mise itself, with `task_config.includes`, and file tasks under `mise-tasks/` |

## How findings are classified

- **False positive:** ci-parity is wrong about the repository. Each one has a bug.
- **Noise:** correct by the rules, but nobody would move it into a task: inline
  `python3 -c` scripts, tool installs pinned to a version (`cargo install --locked
  mdbook@0.4.52`), `--version` and `go env` checks, and registry commands such as
  `docker pull`.
- **True positive:** a project command run inline, such as `cargo test` or `pnpm build`.

Every P002 and every suggestion is checked against the runner files; P001s are classified
by command.

## Results for v0.8.0

| Repository | Runner | P001 | P002 | P005 | False positives | Noise |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| go-task/task | Taskfile | 13 | 2 | 0 | 2 | 2 |
| goreleaser/goreleaser | Taskfile | 4 | 0 | 1 | 0 | 3 |
| arduino/arduino-cli | Taskfile | 39 | 0 | 2 | 23 | 9 |
| prometheus/prometheus | make | 12 | 4 | 1 | 4 | 3 |
| traefik/traefik | make | 20 | 0 | 3 | 0 | 8 |
| cli/cli | make | 20 | 0 | 7 | 0 | 8 |
| casey/just | just | 10 | 0 | 0 | 0 | 2 |
| oxc-project/oxc | just | 121 | 0 | 17 | 0 | 3 |
| jdx/mise | mise | 20 | 1 | 8 | 1 | 7 |
| jdx/hk | mise | 44 | 2 | 14 | 2 | 33 |
| **Total** | | **303** | **9** | **53** | **32** | **78** |

- **False positives: 32 of 312 errors (10.3%).** All 9 P002s were wrong, and 23 of
  arduino-cli's 39 P001s are task calls through `go tool`.
- **Noise: 78 (25%):** 33 inline scripts, 28 tool installs, 13 version checks and 4
  registry commands.
- **True positives: 202 (65%).**
- **P005 warnings:** all 53 are expected: 38 PowerShell steps and 15 command names built
  from expressions or variables.
- **Suggestions: 18, of which 5 run exactly the step.** The other 13 name a task that runs
  more, and `fix` rewrites some of them: casey/just's `cargo fmt --all -- --check` became
  `just ci`.

Each cause has an issue:

| Issue | Cause | Findings |
| --- | --- | ---: |
| [#55](https://github.com/jimmykurian/ci-parity/issues/55) | `go tool …/cmd/task` isn't recognized as calling task | 23 P001 |
| [#54](https://github.com/jimmykurian/ci-parity/issues/54) | make pattern rules such as `%: common-% ;` | 4 P002 |
| [#45](https://github.com/jimmykurian/ci-parity/issues/45) | mise `_default` and `.py` file tasks | 3 P002 |
| [#44](https://github.com/jimmykurian/ci-parity/issues/44) | the values of task's `--output-group-*` flags | 2 P002 |
| [#53](https://github.com/jimmykurian/ci-parity/issues/53) | suggestions and `fix` match a task that runs more than the step | 13 suggestions |
| [#50](https://github.com/jimmykurian/ci-parity/issues/50) | how to handle the noise | 78 P001 |

The 1.0 criteria ([#51](https://github.com/jimmykurian/ci-parity/issues/51)) ask for a
re-run with no false positives.

## Results for v0.10.0

The same commits, after the fixes in the table above and the built-in allow set (#50).

| Repository | Runner | P001 | P002 | P005 | Allowed | False positives | Noise | True |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| go-task/task | Taskfile | 12 | 0 | 0 | 1 | 0 | 1 | 11 |
| goreleaser/goreleaser | Taskfile | 3 | 0 | 1 | 1 | 0 | 2 | 1 |
| arduino/arduino-cli | Taskfile | 12 | 0 | 2 | 4 | 0 | 5 | 7 |
| prometheus/prometheus | make | 12 | 0 | 1 | 0 | 0 | 3 | 9 |
| traefik/traefik | make | 18 | 0 | 3 | 2 | 0 | 6 | 12 |
| cli/cli | make | 12 | 0 | 7 | 8 | 0 | 0 | 12 |
| casey/just | just | 8 | 0 | 0 | 2 | 0 | 0 | 8 |
| oxc-project/oxc | just | 120 | 0 | 17 | 1 | 0 | 2 | 118 |
| jdx/mise | mise | 18 | 0 | 8 | 2 | 0 | 6 | 12 |
| jdx/hk | mise | 11 | 0 | 14 | 33 | 0 | 0 | 11 |
| **Total** | | **226** | **0** | **53** | **54** | **0** | **25** | **201** |

P001 counts only errors here; allowed findings are in their own column.

- **False positives: none.** There are no P002s, and the 23 `go tool` task calls are
  checked as task calls.
- **Allowed: 54** by the built-in allow set: 33 inline scripts, 11 version or environment
  checks and 10 pinned tool installs. Every one is a setup or info command.
- **Noise left: 25 (11% of errors):**
  - 18 tool installs that aren't pinned, such as `npm i -g corepack`,
    `pip install poetry`, `go install …@latest` and `cargo binstall cargo-zigbuild`,
    which the built-in set leaves out on purpose;
  - 1 pinned tool run once, `npx --yes junit-report-merger@9.0.3`;
  - 5 registry commands (`docker buildx imagetools`, `docker pull`);
  - 1 version check run inside `docker run`.
- **True positives: 201 (89%):** builds, tests and checks run inline, such as
  `cargo build`, `go test` and `pnpm --filter … build`.
- **P005 warnings:** the same 53 as before, all expected.
- **Suggestions: 5, all exact:** each names a task that runs only that command.

## See also

- [Testing strategy](testing-strategy.md)
- [P001](../rules/p001.md) and [P002](../rules/p002.md)
- [How runner files are read](../reference/runners.md)

Last reviewed: 2026-10-29
