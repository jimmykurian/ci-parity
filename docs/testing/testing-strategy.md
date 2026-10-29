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
| Performance | `test/perf/` | `check` on a generated repository of 50 workflows (1,200 steps) takes under a second; `task bench` reports time and allocations |
| Schemas | `schemas/`, `task lint:schemas` | Every JSON golden, fixture config and baseline matches its JSON Schema, and the deliberately broken ones don't; a unit test keeps the config schema's keys in step with the parser |
| Field trial | `test/trial/`, `task trial` | ci-parity on 10 public repositories at pinned commits; every P002 and suggestion is checked by hand. Not in `task ci`, since it needs the network. See [Field trial](field-trial.md) |
| Action tests | the **Self-check** and **Action download** CI jobs | The action from this checkout on this repository and on a fixture with findings; and the published action downloading, verifying and running the latest release on Linux, macOS and Windows |

Every test layer runs on Linux in the **Checks** job and on macOS and Windows in the
**Tests** jobs, so path, line-ending and file-mode differences show up in CI.

Unit tests stay in their package so they can test unexported functions, and are plain
`testing` table tests. The integration and end-to-end tests use the program only from
outside, so they live under `test/`, written as [Ginkgo](https://onsi.github.io/ginkgo/)
specs with [Gomega](https://onsi.github.io/gomega/) matchers.

## Ginkgo specs

Each package under `test/` has one `TestX` function that runs its specs, so `go test` and
the `task test:*` commands run them like any other test; the Ginkgo CLI isn't needed.

- Specs read as behavior: `Describe("ci-parity baseline")`, `Context("when the
  inline-tool counts rose")`, `It("refuses, listing the workflow, and leaves the file
  unchanged")`. Table tests are `DescribeTable` with one `Entry` per case.
- Labels: `integration`, `e2e`, and `golden` on the golden table. Select with
  `go test ./test/... -ginkgo.label-filter=golden`; focus on specs by name with
  `-ginkgo.focus='ci-parity fix'`.
- Spec order is randomized, so each spec copies the fixture it writes to.
- The end-to-end suite builds the binary once in `BeforeSuite` with `gexec.Build` and runs
  it with `gexec.Start`, asserting on `gexec.Exit` and `gbytes.Say`.

## Golden fixtures

Each case is a complete miniature repository: config, runner file, workflows and an
optional baseline, plus `expected/{text,github,json,sarif}.golden` and `expected/exitcode`.
The golden table in `test/integration/golden_test.go` has one entry per fixture and
format, built from `testutil.FixtureDirs`, and runs `cli.Run` in process; the comparisons
in `internal/testutil/golden.go` take either a `*testing.T` or `GinkgoT()`. A new fixture
needs no test code.

- Prefer a new fixture to a mock. Use invented names such as `acme-app`, `example-api` and
  `widget-service`.
- `task test -- -update` rewrites the goldens (it runs the `golden` label). Read the diff: every changed line should be
  one you meant.

## Unit tests

Table tests sit next to the code they cover: the shell extractor, the workflow positions,
each runner's reader and call parser, the config globs, suggestions and unwrapping, the
ratchet, `fix`'s edits and the file editor, and the SARIF fields.

Each test, and each `t.Run` body, follows Arrange, Act, Assert, marked with comments and
separated by blank lines:

```go
func TestApply(t *testing.T) {
    // Arrange
    b := &Baseline{Workflows: map[string]int{"a.yml": 2}}
    want := []Row{{"a.yml", 2, 0}}

    // Act
    _, rows := Apply(b, []string{"a.yml"}, nil)

    // Assert
    if diff := cmp.Diff(want, rows); diff != "" {
        t.Errorf("rows (-want +got):\n%s", diff)
    }
}
```

- **Arrange** holds the setup and the expected values; in a table test, the table is the
  Arrange section.
- **Act** is one call to the code under test. A fail-fast `if err != nil { t.Fatal(err) }`
  right after it belongs to Act (gofumpt keeps it against the call); when the error is
  what's being tested, the check is in Assert.
- **Assert** holds the checks. A test that would act twice on unrelated behavior is two
  tests.
- An empty section is left out.

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
- [Field trial](field-trial.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [Releases and CI](../deployment/releases-and-ci.md)

Last reviewed: 2026-10-28
