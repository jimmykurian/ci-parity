<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Security Policy

## About this project

ci-parity is a command-line tool and GitHub Action that reads a repository's
GitHub Actions workflows, composite actions, Taskfiles and its own config, and
reports findings. It never runs the commands it reads, never executes `task`,
and makes no network requests at check time. Its inputs are files in the
repository being checked, which in CI may come from an untrusted pull request.

## Supported versions

ci-parity is pre-1.0. Only the latest `v0.x` release, and `main`, get fixes. A
fix ships as a new patch release, and the floating `v0` tag moves to it.

## Reporting a vulnerability

**Please don't report security issues through public issues or pull
requests.** Use GitHub's private vulnerability reporting: on the repository's
**Security** tab, choose **Report a vulnerability**. If that isn't available,
email <hello@jimmykurian.me>.

Please include:

- The ci-parity version (`ci-parity version`) or the Action ref
- A minimal workflow, Taskfile or config that triggers it
- What an attacker could do with it

### What to expect

This is a one-person project, so these are goals rather than guarantees:

| Action                | Timeframe                         |
| --------------------- | --------------------------------- |
| Acknowledgment        | Within 5 days                     |
| Assessment and update | Within 14 days                    |
| Fix and release       | As soon as practical for severity |

There is no bug bounty. I'm happy to credit you in the release notes once it's
fixed, unless you'd rather stay anonymous.

### In scope

- Input that makes ci-parity execute code, write files other than the
  baseline it was asked to write, or make a network request
- Output injection: a crafted workflow, task name or allow reason that escapes
  the `github` format's workflow-command encoding, or breaks the JSON document
- Crashes, hangs or unbounded memory use on crafted YAML or shell input
- The GitHub Action and release pipeline: a way to run an unverified binary,
  or to publish one

### Out of scope

- Findings ci-parity misses or reports wrongly with no security impact; open a
  "Wrong Finding" issue instead
- Taskfile includes that point outside the repository: like `task` itself,
  ci-parity follows them, and only reads them to list task names
- Vulnerabilities in GitHub Actions, go-task or Go itself: report those to
  their maintainers
- Denial of service against the GitHub runner that needs more than one
  checked repository's own files

## How the project is secured

- **No execution or network at check time:** the code has no `os/exec` or
  network imports outside tests, and runner files are read statically.
- **Few dependencies:** `gopkg.in/yaml.v3`, `mvdan.cc/sh/v3`,
  `github.com/BurntSushi/toml`, and `github.com/google/go-cmp` for tests only.
  Adding one needs agreement in an issue first.
- **Pinned CI:** third-party actions are pinned to full commit SHAs, dev tools
  are pinned in `mise.toml`, and workflows run with `contents: read`.
- **Secret scanning:** gitleaks runs on every push and pull request; a hit
  fails the build.
- **Signed commits:** commits are SSH-signed and show as Verified on GitHub.

## Responsible disclosure

Please give me reasonable time to release a fix before disclosing an issue. I
won't pursue action against anyone reporting in good faith.

## Attribution

Structure adapted from the [OpenSSF Vulnerability Disclosure
Guide][ossf-guide], licensed under [CC-BY-4.0][cc-by].

[ossf-guide]: https://github.com/ossf/oss-vulnerability-guide
[cc-by]: https://creativecommons.org/licenses/by/4.0/
