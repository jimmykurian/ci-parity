<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Configuring

Choose the task runner, adjust which commands count as tools, and allow the inline steps
you mean to keep. Every key is listed with its default in the
[configuration reference](../reference/configuration.md).

## The config file

`.ci-parity.yml` at the repository root is optional, and so is every key in it. These are
the defaults:

```yaml
runner: taskfile                  # or make, just or mise
taskfile: ""                      # empty: Taskfile.yml, taskfile.yml, Taskfile.yaml,
                                  # taskfile.yaml, then the .dist variants
workflows:
  - .github/workflows/*.yml
  - .github/workflows/*.yaml
  - .github/actions/**/action.yml
  - .github/actions/**/action.yaml
invokers:
  shell: [task, go-task, "go tool task"]
  actions:
    - uses: ./.github/actions/run-task
      input: task
tools:
  extend: []                      # added to the built-in list
  remove: []                      # removed from the built-in list
  replace: []                     # non-empty replaces the built-in list
transparent: [sudo, env, time, nice, ionice, timeout, nohup, xvfb-run, chronic]
allow: []
baseline: .ci-parity-baseline.json
```

An unknown key, a wrong type, an `allow` entry without a `reason`, or `replace` combined
with `extend` or `remove` is an error that names the key (exit 2).

## Tools

Only commands on the tools list are flagged; anything else is ignored. Tools are matched
on the command's basename, so `./node_modules/.bin/eslint` and `./gradlew` count. The
built-in list covers:

- node: `npm npx pnpm pnpx pnx yarn bun bunx node deno`
- python: `python python3 pip pip3 pipx uv uvx poetry pytest ruff mypy tox nox`
- `dotnet`, go (`go golangci-lint`) and `cargo`
- jvm: `mvn mvnw gradle gradlew`
- infrastructure: `terraform tofu docker docker-compose kubectl helm az aws gcloud pulumi`
- quality tools: `eslint prettier biome vitest jest playwright cypress k6 shellcheck
  actionlint`
- other task runners: `make just mise cargo-make`, plus `task` and `go-task` when the
  runner isn't Taskfile

The configured runner is never on the list. Add your own with `tools.extend`, stop
flagging one everywhere with `tools.remove`, or bring a whole list with `tools.replace`.

**Transparent** wrappers are looked through: `sudo -E env CI=1 pnpm test` is `pnpm test`.
**Invokers** are how CI calls the runner: shell commands such as `task`, or a local
composite action whose input names the task.

## Runners

`runner` picks the task runner. Its file is read as written; ci-parity never runs it.

| `runner` | Reads | CI calls | Internal tasks ([P003](../rules/p003.md)) |
| --- | --- | --- | --- |
| `taskfile` | `Taskfile.yml` and its `includes` | `task build` | `internal: true` |
| `make` | `GNUmakefile`, `makefile` or `Makefile`, with `include`s | `make build`, `gmake build` | none |
| `just` | `justfile`, `Justfile` or `.justfile`, with `import`s and `mod`ules | `just build`, `just db migrate` | `[private]`, or a name starting with `_` |
| `mise` | `mise.toml` and the other mise config files, `task_config.includes` and file tasks | `mise run build`, `mise r build` | `hide = true`, or `#MISE hide=true` in a file task |

With another runner, messages and suggestions use its syntax (`make lint`, `just lint`,
`mise run lint`), and `task` counts as an inline tool. Some calls aren't checked because
they point elsewhere: make's `-f` or `-C`, just's `-f` or `-d`, mise's `-C` and task
globs such as `lint:*`. The [runners reference](../reference/runners.md) has every
detail, and the [recipes](recipes.md) show a passing setup for each.

## Allowing a step

Some inline commands are deliberate. Allow one with a comment on the same line or the
line above, with a reason:

```yaml
      - run: docker push ghcr.io/acme/widget-service:dev # ci-parity: allow images publish only from CI
```

Or allow a command pattern in the config, optionally for some workflows only:

```yaml
allow:
  - match: "docker push *"           # matched against the whole command
    reason: images publish only from CI
    workflows: [.github/workflows/release.yml]
```

Allowed findings don't fail the check; `--verbose` lists them with their reason. An allow
comment without a reason is reported as [P005](../rules/p005.md).

## See also

- [Configuration reference](../reference/configuration.md)
- [Runners reference](../reference/runners.md)
- [Recipes](recipes.md)

Last reviewed: 2026-10-28
