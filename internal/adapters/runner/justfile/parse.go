// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package justfile

import (
	"regexp"
	"strings"
)

// importLine is an `import 'path'` line; `import?` makes it optional.
type importLine struct {
	path     string
	optional bool
}

// module is a `mod name` line, with its path when one is given; `mod?` makes it
// optional.
type module struct {
	name, path string
	optional   bool
}

// parser reads one justfile's lines into its Justfile, under prefix.
type parser struct {
	j       *Justfile
	prefix  string
	private bool    // a [private] attribute applies to the next recipe or alias
	ownDir  bool    // a [no-cd] or [working-directory] attribute applies to the next recipe
	current *recipe // recipe whose body is being read
	imports []importLine
	modules []module
	added   []*recipe // recipes and aliases this file defines
	// setWorkingDir records a `set working-directory`, which moves every recipe.
	setWorkingDir bool
}

var (
	aliasLine  = regexp.MustCompile(`^alias\s+([A-Za-z_][A-Za-z0-9_-]*)\s*:=\s*([A-Za-z_][A-Za-z0-9_:-]*)\s*$`)
	importRe   = regexp.MustCompile(`^import(\?)?\s+['"]([^'"]+)['"]\s*$`)
	moduleRe   = regexp.MustCompile(`^mod(\?)?\s+([A-Za-z_][A-Za-z0-9_-]*)(?:\s+['"]([^'"]+)['"])?\s*$`)
	attribute  = regexp.MustCompile(`^\[(.*)\]\s*$`)
	recipeHead = regexp.MustCompile(`^@?([A-Za-z_][A-Za-z0-9_-]*)`)
)

// parse reads text line by line.
func (p *parser) parse(text string) {
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		p.line(line)
	}
}

// line reads one line: an indented line belongs to the current recipe's body;
// otherwise it is a comment, an attribute, an alias, an import, a module, or
// anything else, which recipe reads.
func (p *parser) line(line string) {
	if line != "" && (line[0] == ' ' || line[0] == '\t') {
		p.body(line)
		return
	}
	text := strings.TrimSpace(line)
	if text == "" {
		return // a blank line may sit inside a recipe body
	}
	p.current = nil
	if strings.HasPrefix(text, "#") {
		return
	}
	if m := attribute.FindStringSubmatch(text); m != nil {
		for _, a := range strings.Split(m[1], ",") {
			switch a = strings.TrimSpace(a); {
			case a == "private":
				p.private = true
			case a == "no-cd" || strings.HasPrefix(a, "working-directory"):
				p.ownDir = true
			}
		}
		return
	}
	switch {
	case aliasLine.MatchString(text):
		m := aliasLine.FindStringSubmatch(text)
		p.add(m[1], &recipe{alias: p.prefix + m[2]})
	case importRe.MatchString(text):
		m := importRe.FindStringSubmatch(text)
		p.imports = append(p.imports, importLine{path: m[2], optional: m[1] == "?"})
	case moduleRe.MatchString(text):
		m := moduleRe.FindStringSubmatch(text)
		p.modules = append(p.modules, module{name: m[2], path: m[3], optional: m[1] == "?"})
	default:
		p.recipe(text)
	}
	p.private, p.ownDir = false, false
}

// recipe reads a recipe header, `name param... : dependencies`. Assignments
// (`x := y`), settings and exports define no recipe.
func (p *parser) recipe(text string) {
	m := recipeHead.FindStringSubmatch(text)
	if m == nil {
		return
	}
	name := m[1]
	if name == "set" {
		if f := strings.Fields(text); len(f) > 1 && strings.HasPrefix(f[1], "working-directory") {
			p.setWorkingDir = true
		}
		return
	}
	if name == "export" || name == "unexport" {
		return
	}
	rest := text[len(m[0]):]
	colon := colonOutsideQuotes(rest)
	if colon < 0 || strings.HasPrefix(rest[colon:], ":=") {
		return
	}
	r := &recipe{}
	for _, param := range strings.Fields(rest[:colon]) {
		param = strings.TrimPrefix(param, "$")
		switch {
		case strings.HasPrefix(param, "+"):
			r.min++
			r.max = -1
		case strings.HasPrefix(param, "*"):
			r.max = -1
		case strings.Contains(param, "="):
			if r.max >= 0 {
				r.max++
			}
		default:
			r.min++
			if r.max >= 0 {
				r.max++
			}
		}
	}
	p.add(name, r)
	p.current = r
}

// add records a recipe or alias under the module prefix.
func (p *parser) add(name string, r *recipe) {
	r.private = p.private || strings.HasPrefix(name, "_")
	r.ownDir = p.ownDir
	p.j.recipes[p.prefix+name] = r
	p.added = append(p.added, r)
}

// body records a recipe line without the @ and - prefixes that only change
// how just runs it. Comments and shebang lines are not commands.
func (p *parser) body(line string) {
	if p.current == nil {
		return
	}
	cmd := strings.Join(strings.Fields(strings.TrimLeft(strings.TrimSpace(line), "@-")), " ")
	if cmd == "" || strings.HasPrefix(cmd, "#") {
		return
	}
	p.current.commands = append(p.current.commands, cmd)
}

// colonOutsideQuotes returns the index of the first colon not inside a quoted
// default value, or -1.
func colonOutsideQuotes(s string) int {
	var quote byte
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '\'' || c == '"':
			quote = c
		case c == ':':
			return i
		}
	}
	return -1
}
