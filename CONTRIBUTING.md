<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Contributing

Thanks for helping with ci-parity. Bug reports, wrong findings and focused pull
requests are all welcome. Everyone taking part follows the
[Code of Conduct](CODE_OF_CONDUCT.md). Security issues go through
[SECURITY.md](SECURITY.md), not public issues.

## Before you start

ci-parity's behavior is documented as a contract, so most changes start in the docs:

- [docs/reference/](docs/reference/README.md) says exactly what the tool does: the
  rules and their messages, config keys, the baseline format, output formats and exit
  codes. A new config key, rule or dependency, or a change to a rule ID, a message, the
  baseline format or the JSON or SARIF schema, needs agreement in an issue first, and
  updates the reference in the same pull request as the code.
- [docs/architecture/](docs/architecture/README.md) shows how it fits together and why,
  including ideas not yet planned.
- [AGENTS.md](AGENTS.md) has the working rules for this repo, for people and
  coding agents alike.

For anything bigger than a small fix, open an issue first so we can agree on
the approach.

## Getting started

**Prerequisites:**

- Go, at the version in `go.mod` (Go downloads the pinned toolchain itself)
- [mise](https://mise.jdx.dev), which installs task and every linter at the
  versions in `mise.toml`. Without mise, install those yourself at the same
  versions. Linter configs live in `.config/linters/`.

Or open the repository in the devcontainer (`.devcontainer/`, for VS Code or
GitHub Codespaces), which installs Go at the `go.mod` toolchain and every tool
in `mise.toml`, then runs `task build`.

In VS Code, `.vscode/` recommends the extensions (the same list the devcontainer
installs), formats and lints Go on save with gofumpt and golangci-lint, points the
markdownlint and typos extensions at `.config/linters/`, makes `task build` and
`task test` the default build and test tasks, and adds launch configurations to
debug `ci-parity check` on this repository or a fixture, `fix --dry-run`, and the
tests in the current package.

**Setup:**

```sh
git clone https://github.com/jimmykurian/ci-parity.git
cd ci-parity
git config commit.template .gitmessage

mise install
task hooks:install
task ci
```

Every command runs through the [Taskfile](Taskfile.yml), and CI runs the same
tasks:

| Task | What it does |
| --- | --- |
| `task build` | builds `./bin/ci-parity` |
| `task test` | runs every test, `go test ./...` |
| `task test:unit` | the unit tests, which sit next to the code |
| `task test:integration` | the CLI in-process on every fixture (`test/integration/`) |
| `task test:e2e` | builds the binary and runs it as a process (`test/e2e/`) |
| `task test -- -update` | regenerates the golden files |
| `task lint` | every linter below |
| `task lint:go` | golangci-lint |
| `task lint:vuln` | govulncheck: known vulnerabilities in Go dependencies |
| `task lint:actions` | actionlint and zizmor on the workflows |
| `task lint:shell` | shellcheck |
| `task lint:yaml` | yamllint on the repo's YAML (fixtures are left out) |
| `task lint:markdown` | markdownlint-cli2, run with `pnx` |
| `task lint:spelling` | typos |
| `task lint:editorconfig` | editorconfig-checker |
| `task lint:headers` | Every file starts with the license header |
| `task lint:release` | goreleaser check on `.goreleaser.yaml` |
| `task fmt` | formats Go code with gofumpt and goimports; `task fmt:check` prints the diff and fails |
| `task self-check` | runs ci-parity against this repo's own workflows |
| `task ci` | all of the above, in order |
| `task hooks:install` | turns on the git hooks in `.githooks/` |

## How the code is organized

The layout is clean architecture, with dependencies pointing inward:

- `internal/domain`: types and pure logic, standard library only
- `internal/rules`: the rules, P001 to P005
- `internal/app`: the `check`, `baseline` and `fix` use cases and the ports they need
- `internal/adapters/...`: YAML, shell and runner-file parsing, the baseline file
  and the output formats
- `internal/cli`: flags, wiring and exit codes
- `cmd/ci-parity`: `main`, which calls `cli.Run`

## Tests

Tests come in three layers; [the testing strategy](docs/testing/testing-strategy.md)
says what each proves.

- **Unit tests** sit next to the code they test, in the same package, so they can
  reach unexported functions.
- **Integration tests** in `test/integration/` run the CLI in-process on the
  fixtures, written as Ginkgo specs with Gomega matchers. Golden tests drive the project: each case under
  `testdata/fixtures/<case>/` is a complete miniature repo (config, runner file,
  workflows, optional baseline) with `expected/{text,github,json,sarif}.golden`
  and `expected/exitcode`. The harness runs every fixture in every format, so
  adding a fixture needs no test code.
- **End-to-end tests** in `test/e2e/` build the binary and run it as a process,
  as Ginkgo specs using `gexec`.

Write unit tests as `testing` table tests in Arrange, Act and Assert sections, and
integration and end-to-end tests as `Describe` / `Context` / `It` specs; see the testing
strategy for the conventions, labels and focus.

- Prefer a new fixture over a mock. Use invented names like `acme-app`,
  `example-api` or `widget-service`; never copy a real repository's workflows.
- After `task test -- -update`, read the golden diff. Every change in it should
  be one you meant.
- Parsers, readers and helpers get table tests next to the code.

## Commits

Messages follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).
`.gitmessage` lists the types and scopes:

```text
feat(shell): peel transparent wrappers before matching tools
fix(workflow): report block-scalar run lines one line below the key
test(fixtures): add includes-nested fixture
```

Rule IDs, the baseline format and the JSON schema are public contracts:
changing one is a breaking change, marked with `!` or a `BREAKING CHANGE`
footer.

Commits should be signed so GitHub shows them as Verified. With an SSH key:

```sh
git config gpg.format ssh
git config user.signingkey ~/.ssh/id_ed25519.pub
git config commit.gpgsign true
```

Then add the same public key to GitHub as a signing key.

In the devcontainer, a `user.signingkey` that names a key file on your machine
doesn't exist, so commits fail to sign. Either commit from the host, or point
git at the key itself, which signs through the SSH agent VS Code forwards into
the container and works on the host too:

```sh
git config user.signingkey "key::$(cat ~/.ssh/id_ed25519.pub)"  # run on the host
```

## Git hooks

`task hooks:install` points git at `.githooks/`:

- **pre-commit** refuses commits on `main`, then checks what's staged: gitleaks on
  every commit, then golangci-lint (format and lint), editorconfig-checker,
  the license headers, typos, markdownlint, yamllint, shellcheck, actionlint and zizmor on the files they cover.
- **commit-msg** checks the header against Conventional Commits, with the types and
  scopes in `.gitmessage`.
- **pre-push** refuses pushes to `main`, then runs `task test` and `task self-check`
  when the pushed commits touch Go code, fixtures, the module, the Taskfile, the
  workflows or `.ci-parity*`. It runs on the working tree, so it refuses when that
  isn't a clean checkout of what's being pushed.

A check whose tool isn't installed is skipped with a warning; CI runs everything.
`--no-verify` skips the hooks once.

## Pull requests

- `main` takes changes only through pull requests: work on a branch, and the hooks
  refuse commits and pushes on `main`.
- Keep each pull request to one change, and fill in the
  [template](.github/PULL_REQUEST_TEMPLATE.md).
- The PR title follows Conventional Commits.
- `task ci` must pass. CI runs each of its tasks as its own step, plus gitleaks, and a
  gitleaks hit fails the build; remove the secret rather than allowlisting it.
- On every pull request, CI posts one comment with a table of each check's result,
  the commit it ran on and a link to the run, and updates it on each push.
- The PR is labeled by the paths it changes (`.github/labeler.yml`) and assigned to
  whoever opened it; bots' PRs go to the owner. `.github/CODEOWNERS` requests review.
- Once the repository is public, a GitHub ruleset protects `main`: squash merges only,
  signed commits, linear history, and the **Checks** and **Self-check** jobs must pass.
- Every `run:` step in this repo's workflows calls `task`. Setup goes in
  `uses:` steps pinned to a full commit SHA with a version comment. Anything
  else needs a `# ci-parity: allow <reason>` comment on the line.
- Update `README.md` when a command, flag or output changes.
- Output and message changes follow the rules in
  [ACCESSIBILITY.md](ACCESSIBILITY.md): meaning in words, never color or
  symbols alone.

## Releases

[Release Please](https://github.com/googleapis/release-please) (`.github/workflows/release.yml`)
versions the project from the Conventional Commits on `main`:

- After each merge it opens or updates one release PR, `chore: release <version>`, that
  bumps the version in `.github/.release-please-manifest.json` and adds the commits to
  `CHANGELOG.md`. Before 1.0, a `feat` is a minor release and a `fix` a patch; a `!` or
  `BREAKING CHANGE` footer is a minor release too. `docs` and `refactor` commits are
  listed; `chore`, `ci`, `build`, `test` and `style` aren't.
- The workflow uses `GITHUB_TOKEN`, so the release PR doesn't start the CI checks on its
  own. Close and reopen it to run them, then merge it.
- Merging it tags `vX.Y.Z` and publishes the GitHub Release. The same workflow then runs
  goreleaser to attach the archives and `checksums.txt`, and moves the `v0` tag to the
  release (`task release:float`).
- Never edit a version or the changelog by hand. `task release:snapshot` builds every
  archive into `dist/` locally, and `task lint:release` checks `.goreleaser.yaml`.

## License

By contributing, you agree that your contributions are licensed under the
[Apache License 2.0](LICENSE).
