<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Documentation style guide

How to write and maintain the docs in `docs/`, the READMEs and CONTRIBUTING.md.

## Tone and voice

- **Lead with the answer.** Say what something does or what to do, then why.
- **Everyday words.** Name things by what they are, and explain a term the first time
  it appears.
- **Active voice.** "CI uploads the SARIF log", not "the SARIF log is uploaded".
- **Imperative for steps.** "Run `task ci`", not "You should run `task ci`".
- **Facts, checked against the code.** Describe what the code and CI do today, and show
  real output, not invented output. Don't describe plans as if they shipped.

## Document structure

Every document in `docs/` follows this shape:

```markdown
# Title

One line on what this document covers.

## Section

Content.

## See also

- [Related document](../path/to/doc.md)

Last reviewed: YYYY-MM-DD
```

- **One H1**, matching the topic, then a one-line description.
- **"See also"** when the document points to other docs.
- **"Last reviewed"** as the last line, updated whenever the document changes.

Headings are sentence case. File names are lowercase with hyphens
(`testing-strategy.md`), apart from `README.md` and this file.

## The reference is a contract

Pages under `reference/` state exact behavior: rule IDs, message wording, config keys,
the baseline format, and the JSON and SARIF schemas. Change them in the same pull request
as the code, and only after the change has been agreed in an issue or the pull request.
Mark behavior added after v0.1 with the version it arrived in ("from v0.3").

## Formatting

- **Diagrams:** [Mermaid](https://mermaid.js.org/), which GitHub renders in place. Quote
  labels with punctuation, and check a changed diagram renders on GitHub before merging.
- **Code blocks:** give every fence a language: `sh`, `yaml`, `json`, `go`, `text` for
  output and trees, `mermaid` for diagrams.
- **Links:** relative paths to files in the repository; link text that says where it
  goes.
- **Lint:** `task lint:markdown` and `task lint:spelling` check every page.

## See also

- [Documentation index](README.md)
- [ACCESSIBILITY.md](../ACCESSIBILITY.md)

Last reviewed: 2026-10-28
