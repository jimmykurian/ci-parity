// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package justfile

import (
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Command is how CI calls a just recipe.
func (j *Justfile) Command() string { return "just" }

// Internal is how P003 describes a private recipe.
func (j *Justfile) Internal() string { return "is private" }

// otherFile are flags that point just at another justfile or directory, or
// change the file names it looks for.
var otherFile = map[string]bool{
	"-f": true, "--justfile": true, "-d": true, "--working-directory": true,
	"-g": true, "--global-justfile": true, "--justfile-name": true,
}

// valueFlags take the next argument as their value, as just 1.59's options
// define them; --set takes two.
var valueFlags = map[string]int{
	"--set": 2, "--alias-style": 1, "--ceiling": 1, "--chooser": 1, "--color": 1,
	"--command-color": 1, "--cygpath": 1, "--dotenv-command": 1, "--dotenv-filename": 1,
	"-F": 1, "--dotenv-path": 1, "-E": 1, "--dump-format": 1, "--evaluate-format": 1,
	"--global-fallback-justfile": 1, "--group": 1, "--indentation": 1, "--jobs": 1,
	"--list-heading": 1, "--list-prefix": 1, "--shell": 1, "--shell-arg": 1,
	"--tempdir": 1, "--timestamp-format": 1,
}

// subcommands make just do something other than run recipes, so a call with
// one names none.
var subcommands = map[string]bool{
	"-l": true, "--list": true, "--summary": true, "-s": true, "--show": true, "--evaluate": true,
	"--choose": true, "--dump": true, "-e": true, "--edit": true, "--init": true, "--fmt": true,
	"--variables": true, "--completions": true, "--groups": true, "-c": true, "--command": true,
	"--changelog": true, "--man": true, "--eval": true, "--format": true, "--initialize": true,
	"--clean": true, "--json": true, "--request": true, "--usage": true,
}

// Names returns the recipes a just call names, each with the position of its
// first word. The words after a recipe are its arguments, as many as its
// parameters take; `just db migrate` and `just db::migrate` both name the
// module recipe db::migrate. Scanning stops after an unknown recipe, as just
// would, so its arguments are not reported as recipes too. A call that selects
// another justfile, runs a subcommand or names no recipe returns none.
func (j *Justfile) Names(args []domain.Word) []domain.Word {
	var names []domain.Word
	options := true
	for i := 0; i < len(args); i++ {
		a := args[i]
		if options {
			if skip, stop := j.option(a.Text); stop {
				return nil
			} else if skip >= 0 {
				i += skip
				continue
			}
			options = false
		}
		if !a.Literal {
			return names // an expression: the rest cannot be read
		}
		name, used := j.resolve(args[i:])
		if _, recipe := j.recipes[name]; !recipe && j.modules[name] {
			return names // a module alone runs its default recipe
		}
		names = append(names, domain.Word{Text: name, Literal: true, Line: a.Line, Col: a.Col})
		i += used - 1
		r, ok := j.recipes[name]
		if !ok {
			return names
		}
		if r.max < 0 {
			return names // a variadic recipe takes the rest
		}
		i += min(r.max, len(args)-1-i)
	}
	return names
}

// option handles one word before the first recipe. It returns how many
// following words the option consumes, or -1 when the word is not an option
// or override, and stop when the call is not checked at all.
func (j *Justfile) option(arg string) (skip int, stop bool) {
	flag, _, hasValue := strings.Cut(arg, "=")
	switch {
	case arg == "--":
		return 0, false
	case otherFile[flag] || (len(arg) > 2 && !strings.HasPrefix(arg, "--") && (strings.HasPrefix(arg, "-f") || strings.HasPrefix(arg, "-d"))):
		return 0, true
	case subcommands[flag]:
		return 0, true
	case strings.HasPrefix(arg, "-"):
		if n, ok := valueFlags[flag]; ok && !hasValue {
			return n, false
		}
		return 0, false
	case hasValue:
		return 0, false // NAME=VALUE overrides a variable
	}
	return -1, false
}

// resolve reads a recipe name from the start of words, joining module words:
// `db migrate` is db::migrate when db is a module. It returns the name and how
// many words it used.
func (j *Justfile) resolve(words []domain.Word) (string, int) {
	name, used := words[0].Text, 1
	for j.modules[name] && used < len(words) && words[used].Literal {
		next := name + "::" + words[used].Text
		if _, ok := j.recipes[next]; !ok && !j.modules[next] {
			break
		}
		name, used = next, used+1
	}
	return name, used
}
