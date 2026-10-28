# Runners

How ci-parity reads each task runner's files statically, and how calls to each runner are parsed.

## Taskfile

Static parsing of Taskfile v3 (a missing `version` is treated as 3; any other version
exits 2). Collect:

- Task names from `tasks`, plus every entry of each task's `aliases`.
- `internal: true` per task.
- `includes`: a string value is a path to a Taskfile or to a directory (the first of the
  root discovery names found there); a map value supports `taskfile`, `dir`, `optional`, `internal`,
  `aliases`, `flatten`. Included tasks are named `<namespace>:<task>` (also under each
  alias namespace), or unprefixed when `flatten: true`. `internal: true` on an include
  marks all its tasks internal. Resolve relative to the including file's directory, cap
  recursion at depth 10, ignore cycles. A missing include is a parse error (exit 2)
  unless `optional: true`. An include path holding a template (`{{OS}}`) is treated as
  missing.
- An included Taskfile with a `default` task also makes the bare namespace a valid name.
- File names are matched exactly against the directory listing.
- Each task's command strings for suggestion matching: `cmd:` on the task, string
  entries of `cmds`, and `cmd:` map entries of `cmds`. Ignore `task:`, `for:` and
  `defer:` entries.

## Make, just and mise

`runner` selects `taskfile`, `make`, `just` or `mise`. Each has a reader behind the same
interface and its own call syntax; the rules are unchanged.

- Discovery: the root file is found in the repository root, names matched exactly.
  make: `GNUmakefile`, `makefile`, `Makefile`. just: `justfile`, `Justfile`,
  `.justfile`. mise: every one of `.config/mise/config.toml`, `.config/mise.toml`,
  `mise/config.toml`, `.mise/config.toml`, `mise.toml` and `.mise.toml` that exists,
  read in that order, a later file's task replacing an earlier one's; messages name the
  last one found. The `taskfile` key is only for `runner: taskfile`; set with another
  runner it is a config error. A missing root file exits 2.
- Defaults by runner: `invokers.shell` is `[make, gmake]`, `[just]` or
  `["mise run", "mise r", "mise tasks run"]`, and `invokers.actions` is empty, unless
  the config sets them. Unless `tools.replace` is set, the configured runner's commands
  leave the tools list and the other runners' are on it, including `task` and `go-task`.
- make: a target is a rule line outside a recipe; special targets (`.PHONY`), pattern
  rules (`%`), variable assignments (including target-specific ones), `define` blocks
  and names holding `$` define none. `include`, `-include` and `sinclude` of literal
  paths, resolved from the repository root, are followed to depth 10; a missing file is
  skipped, as make may generate it. Recipe lines, without leading `@`, `-` or `+` and
  with whitespace collapsed, are a target's commands. No target is internal. In a call,
  `-f`, `--file`, `--makefile`, `-C` and `--directory` (in any form) select another
  makefile and the call is ignored; `-I --include-dir -o --old-file --assume-old -W
  --what-if --new-file --assume-new` take a value; `-j --jobs -l --load-average
  --max-load` take the next argument only when it is a number; `NAME=value` is a
  variable; a call naming no target runs the default goal and is ignored.
- just: recipes, their parameters, `alias a := b`, `import` of literal paths and `mod`
  (recipes named `mod::recipe`, from `name.just`, `name/mod.just`, `name/justfile` or a
  given path). A recipe is internal when it is `[private]` or its name starts with `_`.
  Body lines without leading `@` or `-` are its commands. In a call, the words after a
  recipe are its arguments, as many as its parameters take (all the rest for a variadic
  one); `just mod recipe` and `just mod::recipe` both name a module recipe. `-f`,
  `--justfile`, `-d` and `--working-directory` select another justfile; `--set` takes
  two values.
- mise: `[tasks.name]` tables and `name = "cmd"` (or list) entries, with `run` (string
  or list) as the commands, `alias` (string or list) as aliases and `hide = true` as
  internal; executable file tasks under `.config/mise/tasks/`, `mise/tasks/`,
  `.mise/tasks/`, `mise-tasks/` and `.mise-tasks/` (named from the path, `db/migrate`
  as `db:migrate`, hidden by a `#MISE hide=true` line, with no commands); and
  `task_config.includes`, each a task TOML file whose top-level tables are tasks or a
  directory of file tasks, a missing one skipped. A call names one task, then its
  arguments; `:::` starts another task. `-C` and `--cd` select another directory; a
  task name holding a glob (`lint:*`) is not checked; `-j --jobs -E --env -t --tool
  --timeout -o --output -s --shell` take a value.

## See also

- [Reference index](README.md)
- [Analysis](analysis.md)
- [Configuration](configuration.md)

Last reviewed: 2026-10-28
