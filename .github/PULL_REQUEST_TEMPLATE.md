# Description

Summarize the change and why it's needed. Link the related issue.

Fixes # (issue)

## Type of change

Delete the options that don't apply.

- [ ] Bug fix: a wrong, missing or misplaced finding, or a crash
- [ ] New behavior within the spec (a rule case, a Taskfile feature, a wrapper)
- [ ] Output change: text, `github` or `json` output, or exit codes
- [ ] Contract change: a rule ID, config key, baseline format or JSON schema
      (needs a `docs/DECISIONS.md` entry first)
- [ ] GitHub Action or release change
- [ ] Build, CI or tooling change
- [ ] Documentation only

## How Has This Been Tested?

- [ ] `task ci` passes (fmt, lint, test, build, self-check)
- [ ] Added or updated a fixture under `testdata/fixtures/` for the behavior,
      with `text`, `github` and `json` goldens and an `exitcode`
- [ ] Golden changes from `task test -- -update` are reviewed and intended
- [ ] Added or updated table tests next to the code for parser or reader changes

## Checklist

- [ ] The PR title follows Conventional Commits, with a scope where one fits
- [ ] Commits are signed (Verified)
- [ ] Fixtures are synthetic, with invented names like `acme-app` or
      `widget-service`
- [ ] No new dependency, config key or rule beyond `docs/SPEC.md`, unless
      `docs/DECISIONS.md` records why
- [ ] Nothing from the Phase 5 backlog in `docs/PLAN.md`
- [ ] Every `run:` step in this repo's workflows calls `task`, or carries a
      `# ci-parity: allow <reason>` comment
- [ ] New third-party actions are pinned to a full commit SHA with a version
      comment
- [ ] `README.md` and `AGENTS.md` are updated if a command, flag or decision
      changed
- [ ] New or changed output follows `ACCESSIBILITY.md`:
      meaning in words, errors that name what and where

## Output

Paste the `ci-parity check` output before and after the change, or mark as N/A.
