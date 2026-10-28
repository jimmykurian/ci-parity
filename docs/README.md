<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Documentation

How ci-parity works, what it reads and reports, and how it is built, tested, secured and
released. The same pages are published, with search, at [https://jimmykurian.github.io/ci-parity](https://jimmykurian.github.io/ci-parity/).

## Documentation index

| Section | Covers |
| --- | --- |
| [Guides](guides/README.md) | Getting started, configuring, adopting with a baseline, fixing findings, CI setup, recipes per runner, why ci-parity, and the FAQ |
| [Rules](rules/README.md) | P001 to P005: what each catches, an example, and how to fix or allow it |
| [Reference](reference/README.md) | The contract: configuration, the baseline, how workflows and runner files are read, the rules and their messages, the CLI and exit codes, the output formats and the GitHub Action |
| [Architecture](architecture/README.md) | The system context, the packages and their ports, the check, fix and release flows, and the design decisions |
| [Releases and CI](deployment/releases-and-ci.md) | The CI jobs, the check table and badges, the `main` ruleset, and how a release and the Marketplace listing are published |
| [Testing](testing/testing-strategy.md) | Golden fixtures, fix fixtures, unit and use-case tests, and the action's tests |
| [Security](security/overview.md) | The threat model, the action's input handling, the supply chain and the repository controls |

## Related documentation

| File | Covers |
| --- | --- |
| [README.md](../README.md) | What ci-parity is, the status badges, a quick start and the architecture diagram |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Setup, every task, tests, commits and signing, hooks, pull requests and releases |
| [AGENTS.md](../AGENTS.md) | Rules for coding agents |
| [ACCESSIBILITY.md](../ACCESSIBILITY.md) | The accessibility goal for output and docs, known gaps and the rules for changes |
| [SECURITY.md](../SECURITY.md) | How to report a vulnerability, and what's in scope |
| [CHANGELOG.md](../CHANGELOG.md) | Every release, written by Release Please; also published on the site |

## Contributing to the docs

Follow the [style guide](STYLE_GUIDE.md).

Last reviewed: 2026-10-28
