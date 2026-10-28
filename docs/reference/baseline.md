# Baseline

The `.ci-parity-baseline.json` file and how its counts tolerate, lock and ratchet inline-tool findings.

## The file and its rules

```json
{
  "version": 1,
  "workflows": {
    ".github/workflows/ci.yml": 7,
    ".github/workflows/release.yml": 0
  }
}
```

Semantics:

- No baseline file: strict mode. Every P001 finding is an error.
- Workflow present with count N greater than zero: the first N P001 findings in that
  workflow (file order, `allowed` findings excluded) are `tolerated`; the rest are
  `error`, and one P004 finding per workflow reports current versus baseline.
- Workflow present with count 0, or absent: locked. Every P001 finding is an error, and
  one P004 finding reports the lock.
- P002, P003 and P005 are never affected by the baseline.
- `ci-parity baseline` writes the current non-allowed P001 count for every scanned
  workflow, writing `0` explicitly so locks are visible in diffs. It refuses with exit 1
  if any count would rise, listing the workflows, unless `--allow-regression` is given.
- `ci-parity check` prints a ratchet table whenever a current count is below its
  baseline, suggesting `ci-parity baseline` to lock in the improvement. Not an error.
- `--strict` on `check` ignores the baseline entirely.

## See also

- [Reference index](README.md)
- [Rules](rules.md)
- [CLI](cli.md)

Last reviewed: 2026-10-28
