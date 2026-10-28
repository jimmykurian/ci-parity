<!-- Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE. SPDX-License-Identifier: Apache-2.0 -->

# Accessibility

ci-parity is a command-line tool and a GitHub Action, so accessibility here means
output that works with a screen reader, a braille display, screen magnification
and other tools, and documentation that reads well with them. This file sets the
goal, records where the project stands, and lists the rules that keep it there.

## The goal

Everything ci-parity prints can be understood without seeing color, layout or
symbols, and every result is also available in a structured form that other
tools can present however their users need.

For criteria, the project follows the W3C's [guidance on applying WCAG to
non-web software][wcag2ict] where it fits a command-line tool: meaning that
doesn't depend on color or sensory cues (1.3.3, 1.4.1), text that can be read
in order (1.3.2), and errors that are identified and described in text (3.3.1).

## Where it stands

| Area | Status |
| --- | --- |
| Color | None: ci-parity prints no color or ANSI escape codes, on a terminal or off one |
| Severity | Carried by words: `error`, `warning`, `[tolerated]`, `[allowed: <reason>]`, `REGRESSION`, and a summary line with counts |
| Symbols | None in its own text: no icons, box drawing, spinners or progress bars. File paths, step names and commands from your repository are printed as they are |
| Motion and timing | None: no animation, no prompts, nothing that waits for input or times out |
| Structured output | `--format json` gives every finding with all its fields, for editors, dashboards or other tools to present; `--format github` puts findings on the lines they refer to in the pull request |
| Results for scripts | Exit codes `0` (clean), `1` (findings) and `2` (usage or input error) |
| Errors | Written in full sentences on stderr, prefixed `ci-parity:`. A bad config key is named (`invokers.actions[0] needs both uses and input`) |
| Docs | Plain Markdown: headings in order, tables with header rows, no images |

## Known gaps

1. **Text output groups findings under a file heading.** A finding line holds
   the position, rule and message but not the file path, so a screen reader
   user who jumps to a single line, or a `grep` for one, loses which file it's
   in. Workaround: `--format json` or `--format github`, which repeat the file
   on every finding. Changing the text layout is a contract change
   ([output formats](docs/reference/output-formats.md)).
2. **Long lines.** Text output doesn't wrap, and a line with a long command can
   pass 150 characters. Terminals and screen readers wrap it, but magnified
   views may need horizontal scrolling.
3. **No record of manual checks.** The checklist below hasn't been run and
   recorded for a release yet.

## Rules for changes

These apply to every change to output, messages or docs:

- **Never convey meaning by color, position or a symbol alone.** Any new state
  gets a word in the text output. Adding color would be a spec change; if it
  ever happens it must be off unless on a terminal, honor
  [`NO_COLOR`](https://no-color.org), and only repeat what the words say.
- **Messages are sentences.** Name the step, job, file and command in words, as
  the existing rules do (`step "Test" in job "build" runs 'pnpm test' directly`).
  Spell out what to do next where there is a clear next step, as P001's
  suggestion and the baseline refusal do.
- **Errors say what and where.** Name the file, key or flag at fault. Usage and
  input errors go to stderr with exit code `2`; findings go to stdout.
- **No interaction.** ci-parity never prompts or waits; every choice is a flag.
- **Keep the JSON complete.** Anything the text format shows is in the JSON
  document, so other tools can present it differently.
- **Docs:** headings in order, tables with a header row, a language on every
  code block, link text that says where it goes, and alt text on any image.

## Manual checklist

Run before each release, and note the date and any findings in the release
notes:

- [ ] Read `ci-parity check --verbose` output for a fixture with errors,
      warnings, tolerated and allowed findings with a screen reader (VoiceOver,
      NVDA or Orca), and confirm each line's severity is spoken
- [ ] Read the same output at 200% zoom in a terminal 80 columns wide
- [ ] Check the annotations on a pull request in GitHub's "Files changed" view
- [ ] Check that every new message reads as a sentence on its own

## Reporting a problem

Open an issue: a bug report for output that's wrong or unreadable, or a
feature request with "Accessibility of output or docs" as the area. Or email
<hello@jimmykurian.me>.

[wcag2ict]: https://www.w3.org/TR/wcag2ict-22/
