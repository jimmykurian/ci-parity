<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Compatibility

What ci-parity keeps stable between releases, what may change, and how a change is
announced. It follows [semantic versioning](https://semver.org).

## Before 1.0

Until v1.0.0, a minor release may still change anything on this page; the
[changelog](https://github.com/jimmykurian/ci-parity/blob/main/CHANGELOG.md) lists every
change. From 1.0, the promises below hold.

## Stable

Changing any of these needs a new major version:

- **Rules:** the IDs `P001` to `P005`, their names, their severities, and the wording of
  their [messages](rules.md). People search for messages, so a message only changes with
  a major version.
- **Configuration:** every key in [`.ci-parity.yml`](configuration.md), its meaning and
  its default. The [config schema](https://jimmykurian.github.io/ci-parity/schemas/config.schema.json)
  describes them.
- **Baseline:** the [file format](baseline.md), version 1, and its model: a count per
  workflow that tolerates that many inline-tool findings and only goes down.
- **Output:** the JSON format (its `version`, keys and value types, described by the
  [output schema](https://jimmykurian.github.io/ci-parity/schemas/output.schema.json)),
  the SARIF format (rule IDs, levels, suppressions, fingerprints and `helpUri` pattern),
  and the `github` format's workflow commands.
- **Command line:** the commands `check`, `baseline`, `fix` and `version`, their flags, and
  the [exit codes](cli.md).
- **GitHub Action:** the [inputs and outputs](action.md), and the `v<major>` tag following
  the newest release of that major version.

## Not covered

These can change in any release:

- The text format's spacing and layout. Use `--format json` to parse results.
- Which task a suggestion names when several tasks run the same command.
- New findings. A new rule, a new runner, or a tool added to the built-in list can report
  steps that passed before; this arrives in a minor release, and the changelog says so.
- Anything under `internal/`. ci-parity is a command and an action; it has no Go API.

## Deprecation

A config key, flag or action input that will be removed is first deprecated for at least
one minor release: it keeps working, ci-parity prints a warning on stderr naming its
replacement, and the changelog lists it. It is removed only in the next major release.

## Platforms

Release binaries are built for linux, darwin and windows, each on amd64 and arm64. CI runs
every test on Linux, macOS and Windows, and the action's release download on all three;
output is the same on each.

## See also

- [Reference index](README.md)
- [Releases and CI](../deployment/releases-and-ci.md)

Last reviewed: 2026-10-28
