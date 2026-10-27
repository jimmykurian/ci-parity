# Decisions

Changes to the decisions in `AGENTS.md` or the contracts in `docs/SPEC.md`. Each entry
states the problem, the proposed change, and its status (proposed, accepted, rejected).
Proposed entries block work until the owner answers.

## Log

### 2026-10-27: wrapper flags that take a value (accepted)

Problem: SPEC 5.3 dropped a wrapper's options but not the values of options that take
one, so `sudo -u deploy docker push` read `deploy` as the command and
`timeout -k 5 30s pnpm test` read `30s`. Both produced no finding.

Change: SPEC 5.3 lists, per built-in wrapper, the flags that consume the next argument.
`env -S` hides the command in a string, so that call is dropped. Wrappers added through
`transparent` in config keep the generic rule; no new config key.

### 2026-10-27: Phase 1 clarifications written into the spec (accepted)

- Unnamed steps read `step 4 in job "web"` (SPEC 6).
- The default `allow` list is empty; the old example was a no-op because `gh` is not a
  tool (SPEC 4.2).
- Text output: singular `1 error` / `1 warning`, column widths, and the `[tolerated]` and
  `[allowed: <reason>]` markers under `--verbose` (SPEC 8).
- Positions are exact for folded and multi-line plain values and after `${{ }}`
  expressions; parse errors are reported at the parser's position (SPEC 5.1, 5.3).
- A YAML line comment after a single-line `run:` value is kept as a shell comment, so
  one-line allow comments work (SPEC 5.1).

### 2026-10-27: Phase 2 spec gaps (accepted)

Approved by the owner and applied to SPEC 5.4 to 5.6. Each item says why the spec as
written would give wrong results.

1. SPEC 5.4, shell invokers. Add: "`-t`, `--taskfile`, `-d`, `--dir` (in any form,
   including `--dir=web` and `-dweb`), `-g` and `--global` select a Taskfile other than
   the root one, so the call is ignored." Why: `task -d web build` names a task in
   `web/Taskfile.yml`; checking it against the root Taskfile reports a false P002.
2. SPEC 5.4, both invokers. Add: "A task name that is not a literal word, or a word in an
   action input that holds `${{ }}`, cannot be checked and is skipped." Why: without it,
   `with: {task: "${{ matrix.task }}"}` splits into `${{`, `matrix.task`, `}}` and reports
   `matrix.task` as unknown.
3. SPEC 5.5, includes. Add: "An included Taskfile with a `default` task also makes the
   bare namespace a valid name." Why: `task docs` runs `docs:default`; without this it is
   a false P002.
4. SPEC 5.5, includes. Replace "(meaning `<dir>/Taskfile.yml`)" with "(the first of the
   root discovery names found there)", and add "An include path holding a template
   (`{{OS}}`) is treated as missing" and "File names are matched exactly against the
   directory listing." Why: task itself searches all names in a directory; templates
   cannot be resolved statically; case-insensitive file systems otherwise report a
   different file name than Linux, which breaks deterministic output.
5. SPEC 5.6. Replace "preferring non-internal tasks" with "never suggesting internal
   tasks". Why: `task` refuses to run an internal task from the command line, so
   suggesting one is wrong advice.
6. SPEC 5.6, P002. Replace the rule with: "A task name that has the unknown name as a
   prefix wins, shortest first. Otherwise suggest the name with the lowest Levenshtein
   distance, when that distance is at most 3 and less than the length of the unknown
   name." Why: the spec does not rank its two kinds of candidate, and plain lowest
   distance suggested `ds` for `test` over `test:unit`.
7. SPEC 5.6, P001. Add: "Ties go to the shorter name, then the alphabetically first.
   Only the canonical task name is suggested, not an alias."

### 2026-10-27: Phase 3 readings of SPEC 4.3 and 8 (accepted, SPEC text unchanged)

Gaps the spec leaves open, filled within its wording:

- Composite actions are scanned files, so they get baseline entries like workflows.
- `ci-parity baseline` with no existing file writes every count without refusing. With a
  file, a workflow missing from it counts as locked at 0, so a new workflow with inline
  tools is a regression. Workflows no longer scanned are dropped on rewrite.
- `--baseline` naming a missing file means strict mode, the same as no file.
- The text ratchet block lists improvements and regressions, as the SPEC 8 example does.
- JSON: `action` is always present (null outside composite actions), `step` is null for
  unnamed steps, `ratchet` lists every scanned file when a baseline applies and is null
  otherwise, and P002's `suggestion` is `task <name>` like P001's.
