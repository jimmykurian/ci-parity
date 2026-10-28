# Testing strategy

The test suites, what each one proves, and how to add to them. `task test` runs them all;
`task ci` runs them with every linter, the build and the self-check.

## The layers

| Layer | Where | Proves |
| --- | --- | --- |
| Golden fixtures | `testdata/fixtures/<case>/` | The CLI's exact output and exit code for a whole repository, in every format |
| Fix fixtures | `testdata/fix/<case>/` | `ci-parity fix` changes exactly the expected bytes, writes nothing on a dry run, and is idempotent |
| Unit tests | next to the code, `*_test.go` | Parsers, readers, call parsing, suggestions, the ratchet, edits and reporters, case by case |
| Use-case tests | `internal/app/app_test.go` | `check` and `baseline` against in-memory ports: strict mode, a missing baseline, refused and allowed regressions |
| Action tests | the **Self-check** CI job | The action itself, on this repository and on a fixture with findings |

## Golden fixtures

Each case is a complete miniature repository: config, runner file, workflows and an
optional baseline, plus `expected/{text,github,json,sarif}.golden` and `expected/exitcode`.
The harness (`internal/testutil/golden.go`) runs the CLI in process against every fixture
in every format, so a new fixture needs no test code.

- Prefer a new fixture to a mock. Use invented names such as `acme-app`, `example-api` and
  `widget-service`.
- `task test -- -update` rewrites the goldens. Read the diff: every changed line should be
  one you meant.

## Unit tests

Table tests sit next to the code they cover: the shell extractor, the workflow positions,
each runner's reader and call parser, the config globs, suggestions and unwrapping, the
ratchet, `fix`'s edits and the file editor, and the SARIF fields.

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
