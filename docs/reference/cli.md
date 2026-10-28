<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# CLI

The `ci-parity` commands, their flags, and the exit codes.

## Commands

```text
ci-parity check    [--config PATH] [--baseline PATH] [--format text|github|json|sarif]
                   [--strict] [--verbose] [-C DIR]
ci-parity baseline [--config PATH] [--baseline PATH] [--allow-regression] [-C DIR]
ci-parity fix      [--config PATH] [--dry-run] [-C DIR]
ci-parity version
```

- `--config` and `--baseline` override the config path and the baseline path from the
  config. `-C` changes to DIR before doing anything.
- `--verbose` includes tolerated and allowed findings in text output. JSON always
  includes them.
- `version` prints `ci-parity <version> (<commit>, <date>)`, with values injected by
  the build via `-ldflags`.
- `fix` (from v0.4) rewrites a P001 finding to its suggestion when that cannot change
  what the step does: its `run:` value sits on one line (plain, single- or
  double-quoted, or a literal block with one line); anything before the command is
  transparent wrappers with their options or `NAME=value` assignments (from v0.5), with
  no operator, redirect, quote, expansion, `${{ }}` or comment; from the command to the
  end of the value is exactly the command; a suggestion exists; and the finding is not
  allowed. Only the command and its arguments are replaced, so wrappers, assignments,
  quotes, comments and the rest of the file are unchanged (`sudo pnpm lint` becomes
  `sudo task lint`); an edit whose text no longer matches the file, such as an escaped
  quoted value, is skipped and counted as needing a manual change. It prints `<file>:<line>:<col> '<old>' -> '<new>'` per edit, then
  `fixed <n> steps in <m> files; <k> inline-tool findings need a manual change`, and,
  when a baseline file exists, a hint to run `ci-parity baseline`. `--dry-run` prints
  `would fix` and writes nothing. It exits 0, or 2 on a usage or input error.

## Exit codes

- 0: no error-severity findings.
- 1: at least one error-severity finding, or `baseline` refused a regression.
- 2: usage error, config error, unreadable or unparseable YAML, missing Taskfile.

## See also

- [Reference index](README.md)
- [Output formats](output-formats.md)
- [Baseline](baseline.md)

Last reviewed: 2026-10-28
