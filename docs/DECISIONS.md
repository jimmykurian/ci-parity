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
