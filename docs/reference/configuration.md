<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Configuration

Which files ci-parity scans, and every key in `.ci-parity.yml` with its default.

## Files scanned

- `.github/workflows/*.yml` and `*.yaml`
- `.github/actions/**/action.yml` and `action.yaml` (local composite actions; their
  `runs.steps` are scanned like job steps)

Paths are reported relative to the repo root with forward slashes on all platforms.

Path globs (`workflows`, `allow[].workflows`) match repo-relative forward-slash paths.
`*` matches any characters except `/`, `?` matches one character except `/`, and a whole
`**` segment matches zero or more path segments. No other syntax (no braces, no
character classes). The standard library has no `**` support and no glob library is
allowed, so `internal/domain` implements this matcher itself, with table tests.

## The config file

Every key is optional. Defaults are shown. The [config schema](https://ci-parity.is-a.dev/schemas/config.schema.json)
describes every key, so an editor with a YAML language server can validate the file and
complete keys; add this first line:

```yaml
# yaml-language-server: $schema=https://ci-parity.is-a.dev/schemas/config.schema.json
```

```yaml
runner: taskfile                  # or make, just, mise (from v0.3; see runners.md)
taskfile: ""                      # empty = auto-discover: Taskfile.yml, taskfile.yml,
                                  # Taskfile.yaml, taskfile.yaml, then the .dist variants
workflows:
  - .github/workflows/*.yml
  - .github/workflows/*.yaml
  - .github/actions/**/action.yml
  - .github/actions/**/action.yaml
invokers:
  shell: [task, go-task, "go tool task", "go tool github.com/go-task/task/v3/cmd/task",
          "go run github.com/go-task/task/v3/cmd/task"]
  actions:
    - uses: ./.github/actions/run-task
      input: task
tools:
  extend: []                      # added to the built-in list
  remove: []                      # removed from the built-in list
  replace: []                     # non-empty replaces the built-in list entirely
transparent: [sudo, env, time, nice, ionice, timeout, nohup, xvfb-run, chronic]
allow: []                         # no commands are allowed by default
baseline: .ci-parity-baseline.json
```

An `allow` entry looks like this:

```yaml
allow:
  - match: "docker push *"        # glob matched against the normalized command
    reason: "images publish only from CI"   # required
    workflows: []                 # optional path globs; empty means every workflow
```

`allow[].match` is a command glob, not a path glob: it is matched against the whole
normalized command ([analysis](analysis.md#commands), step 6), `*` matches any run of characters including `/` and
spaces, and `?` matches exactly one character. `docker push *` therefore matches
`docker push ghcr.io/acme/widget-service:1.0.0`. Matching is case-sensitive.

Built-in tools, matched on the command's basename so `./node_modules/.bin/eslint` and
`./gradlew` match:

- node: `npm npx pnpm pnpx pnx yarn bun bunx node deno`
- python: `python python3 pip pip3 pipx uv uvx poetry pytest ruff mypy tox nox`
- dotnet: `dotnet`
- go: `go golangci-lint`
- rust: `cargo`
- jvm: `mvn mvnw gradle gradlew`
- infra: `terraform tofu docker docker-compose kubectl helm az aws gcloud pulumi`
- quality: `eslint prettier biome vitest jest playwright cypress k6 shellcheck actionlint`
- other runners, which count as bypasses when they are not the configured runner:
  `make just mise cargo-make`

Config errors (unknown key, wrong type, `allow` entry without `reason`, `replace` used
together with `extend` or `remove`) exit 2 with a message that names the key.

## See also

- [Reference index](README.md)
- [Runners](runners.md)
- [Baseline](baseline.md)

Last reviewed: 2026-10-28
