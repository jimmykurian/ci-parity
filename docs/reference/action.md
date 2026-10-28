# GitHub Action

The composite action in `action.yml`: its inputs, outputs and steps.

## Inputs, outputs and steps

`action.yml` at the repo root (required for the Marketplace), `runs.using: composite`.

Inputs:

- `version`: release tag to run. Default empty, which resolves to `github.action_ref`
  when that ref is an exact `vX.Y.Z` tag, otherwise to the latest release.
- `binary`: path to a prebuilt `ci-parity`. When set, no download happens. This repo's
  own CI uses it to test the action before a release exists.
- `config`, `baseline`: passed through.
- `format`: default `github`. With `sarif` the log is written to
  `$RUNNER_TEMP/ci-parity.sarif` instead of printed, and the `sarif-file` output holds
  its path; other formats leave that output empty.
- `strict`: default `false`.
- `working-directory`: default `.`.
- `fail-on-error`: default `true`. When `false`, the step succeeds and exposes outputs.

Outputs: `errors`, `warnings`, `tolerated`, read from a JSON run written to
`$RUNNER_TEMP`, and `sarif-file`.

Steps: resolve the version; download `ci-parity_<version>_<os>_<arch>.tar.gz` (zip on
windows) and `checksums.txt` from the release with `gh release download`, using
`GH_TOKEN: ${{ github.token }}`; verify with sha256; extract to `$RUNNER_TEMP/ci-parity`;
run `check` in `working-directory` once with `--format json` to a file for outputs and
once with the requested format for the log; honor `fail-on-error`.

Branding: `icon: git-merge`, `color: blue`.

## See also

- [Reference index](README.md)
- [CLI](cli.md)
- [Releases and CI](../deployment/releases-and-ci.md)

Last reviewed: 2026-10-28
