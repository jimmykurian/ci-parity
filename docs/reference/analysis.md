# Analysis

How ci-parity reads a workflow: its steps, their shells, the commands in each script, and calls to the task runner.

## Steps

Decode each file with `yaml.v3` into a `yaml.Node` tree so every value has a line and
column. For workflows, walk `jobs.<id>.steps[]`. For composite actions, walk
`runs.steps[]`. Steps with `uses:` are checked against action invokers ([Invokers](#invokers)) and are
otherwise ignored. Steps with `run:` are analyzed.

For a `run:` scalar in literal or folded block style, the script's first line is the
node's line plus one; for a plain or quoted scalar it is the node's line. Report the
line of the command inside the script. Lines and columns are exact for plain, quoted,
literal and folded values. For a folded value, rebuild the folding from the source
lines to map each script position back to the source line it came from; if the rebuilt
text does not equal the parsed value, fall back to counting lines from the first one.
Positions found after the `${{ }}` replacement below are mapped back to the original
script text, so an expression earlier on the line does not shift the column.

On a plain or quoted single-line `run:` value, ` # ...` after the value is a YAML
comment that YAML strips from the value. Append it to the script before shell parsing
so `run: docker push acme # ci-parity: allow <reason>` works.

Replace every `${{ ... }}` expression in the script with the token `__GHA_EXPR_<n>__`
before shell parsing, where `<n>` is the expression index within the step.

## Shells

Resolution order: step `shell`, job `defaults.run.shell`, workflow `defaults.run.shell`,
then the OS default. The OS default is `pwsh` when the job's `runs-on` (string or list)
contains `windows`, otherwise `bash`. If `runs-on` contains an expression, assume `bash`.
Composite action steps must declare `shell`; read it. A composite `run:` step without
`shell` produces one P005 finding for the step and the step is skipped.

Analyze when the shell is `bash`, `sh`, or starts with `bash ` or `sh `. Any other shell
(`pwsh`, `powershell`, `cmd`, `python`, custom) produces one P005 finding for the step
and the step is skipped.

## Commands

1. Parse the script with `mvdan.cc/sh/v3/syntax` using `syntax.LangBash` and
   `KeepComments(true)`. A parse error produces one P005 finding for the step, located
   at the error's position mapped back into the file, and the step is skipped. The
   message carries the parser's text without its script-relative position.
2. Walk the tree. For every `*syntax.CallExpr`, ignore its assignments and take the
   first argument. If the word is not a plain literal (it contains an expansion or an
   `__GHA_EXPR_` token), emit P005 "dynamic command name" and continue.
3. Peel transparent wrappers: while the command's basename is in `transparent`, drop it
   along with its options, then re-read the command. Options are arguments that start
   with `-` or contain `=`. A flag listed below for that wrapper also drops the argument
   after it; `--flag=value` and attached short forms (`-uroot`) carry their own value.
   For `timeout` and `nice`, one numeric argument after the options (`30`, `30s`) is
   dropped too. Wrappers added through config get the generic rule only.

   | Wrapper    | Flags that take a separate value |
   |------------|----------------------------------|
   | `sudo`     | `-u -g -h -p -C -D -R -r -t -T -U --user --group --host --prompt --close-from --chdir --chroot --role --type --command-timeout --other-user` |
   | `env`      | `-u -C -S --unset --chdir --split-string` |
   | `time`     | `-f -o --format --output` |
   | `nice`     | `-n --adjustment` |
   | `ionice`   | `-c -n -p -P -u --class --classdata --pid --pgid --uid` |
   | `timeout`  | `-k -s --kill-after --signal` |
   | `xvfb-run` | `-n -s -f -e -p -w --server-num --server-args --auth-file --error-file --xauth-protocol --wait` |

   `env -S`/`--split-string` hides the command inside a string, so the call is dropped
   once its value is consumed, as is a wrapper left with no command (`sudo -v`).
4. Take the basename of the command.
5. If it begins an invoker pattern ([Invokers](#invokers)), extract task names and continue to rule
   evaluation for P002 and P003.
6. Else if it is on the tools list, record a P001 candidate with the normalized command
   (words joined by single spaces, literal words only).
7. Else ignore.

Calls inside `if`, `for`, `while`, `case`, pipelines, `&&`/`||` lists, subshells and
function bodies are all walked. Heredoc bodies are not analyzed.

Allow comments: a comment of the form `# ci-parity: allow <reason>` on the same line as
the command or on the line directly above it marks that command `allowed`. A missing
reason emits P005 "allow comment without reason" and the command is not allowed.

## Invokers

Shell invokers are word sequences. `go tool task` matches the three words `go`, `tool`,
`task`. After the invoker words, parse arguments:

- Stop at `--`.
- Skip flags. These flags consume the next argument: `-t --taskfile -d --dir
  -C --concurrency -I --interval -o --output`. A flag written as `--flag=value`
  carries its value and consumes nothing further. Any other argument starting with `-`
  is a flag that consumes nothing.
- Arguments of the form `KEY=VALUE` are task variables, not task names.
- `-t`, `--taskfile`, `-d`, `--dir` (in any form, including `--dir=web` and `-dweb`),
  `-g` and `--global` select a Taskfile other than the root one, so the call is ignored.
- Every remaining argument is a task name. A task name that is not a literal word cannot
  be checked and is skipped. Zero task names (for example `task --list`) means the call
  is ignored.

Action invokers: a `uses:` step matches when its value equals the configured `uses`
exactly, or equals it followed by `@<ref>`. Task names come from `with.<input>`, split on
whitespace; a word that holds `${{ }}` cannot be checked and is skipped. A matching step
without that input is ignored.

## See also

- [Reference index](README.md)
- [Runners](runners.md)
- [Rules](rules.md)

Last reviewed: 2026-10-28
