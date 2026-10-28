<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Testing strategy

The test suites, what each one proves, and how to add to them. `task test` runs them all;
`task test:unit`, `task test:integration` and `task test:e2e` run one layer, as CI does;
`task ci` runs them with every linter, the build and the self-check.

## The layers

| Layer | Where | Proves |
| --- | --- | --- |
| Unit tests | next to the code, `*_test.go` | Parsers, readers, call parsing, suggestions, the ratchet, edits, reporters and flag handling, case by case |
| Use-case tests (unit) | `internal/app/app_test.go` | `check` and `baseline` against in-memory ports: strict mode, a missing baseline, refused and allowed regressions |
| Integration: golden fixtures | `test/integration/golden_test.go`, `testdata/fixtures/<case>/` | The CLI's exact output and exit code for a whole repository, in every format |
| Integration: baseline and fix | `test/integration/`, `testdata/fix/<case>/` | `baseline` writes and refuses as it should; `fix` changes exactly the expected bytes, writes nothing on a dry run, and is idempotent |
| End-to-end | `test/e2e/` | The built binary, run as a process: the stamped version, exit codes, stdout and stderr, the working directory in place of `-C`, and `fix` writing to disk |
| Action tests | the **Self-check** CI job | The action itself, on this repository and on a fixture with findings |

Unit tests stay in their package so they can test unexported functions; the integration
and end-to-end tests use the program only from outside, so they live under `test/`.

## Golden fixtures

Each case is a complete miniature repository: config, runner file, workflows and an
optional baseline, plus `expected/{text,github,json,sarif}.golden` and `expected/exitcode`.
The harness (`internal/testutil/golden.go`), called from `test/integration`, runs
`cli.Run` in process against every fixture in every format, so a new fixture needs no test
code.

- Prefer a new fixture to a mock. Use invented names such as `acme-app`, `example-api` and
  `widget-service`.
- `task test -- -update` rewrites the goldens. Read the diff: every changed line should be
  one you meant.

## Unit tests

Table tests sit next to the code they cover: the shell extractor, the workflow positions,
each runner's reader and call parser, the config globs, suggestions and unwrapping, the
ratchet, `fix`'s edits and the file editor, and the SARIF fields.

## End-to-end tests

`test/e2e` builds `cmd/ci-parity` once, with a known version stamped in through `-ldflags`,
and runs the binary as a separate process. It covers what in-process tests can't: the
version stamp, real exit codes, which output goes to stdout and which to stderr, running
from the repository directory without `-C`, and `fix` rewriting a copied fixture on disk.
Keep it small: behavior belongs in the integration tests.

## The action

The **Self-check** job runs `uses: ./` with the binary just built, once on this repository
and once on the `inline-tools` fixture with `fail-on-error: false`, then checks its outputs
with `task action:expect -- 8 0 0`. That second run covers the findings path, which this
repository, with no findings, never reaches.

## See also

- [Documentation index](../README.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [Releases and CI](../deployment/releases-and-ci.md)

Last reviewed: 2026-10-28
