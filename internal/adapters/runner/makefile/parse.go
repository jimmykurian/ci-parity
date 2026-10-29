// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package makefile

import "strings"

// parser walks a Makefile's logical lines, tracking the rule whose recipe the
// tab-indented lines belong to.
type parser struct {
	m       *Makefile
	current []string // targets of the rule being read; nil outside a rule
	define  bool     // inside define ... endef
}

// logicalLines splits text into lines and joins backslash continuations.
func logicalLines(text string) []string {
	var out []string
	var cur strings.Builder
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.HasSuffix(raw, "\\") {
			cur.WriteString(strings.TrimSuffix(raw, "\\"))
			cur.WriteByte(' ')
			continue
		}
		cur.WriteString(raw)
		out = append(out, cur.String())
		cur.Reset()
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// conditionals are directives that may sit inside a recipe without ending it.
var conditionals = map[string]bool{"ifeq": true, "ifneq": true, "ifdef": true, "ifndef": true, "else": true, "endif": true}

// line handles one logical line. It returns the file names of an include
// directive for the caller to read.
func (p *parser) line(line string) ([]string, bool) {
	if p.define {
		if strings.TrimSpace(line) == "endef" {
			p.define = false
		}
		return nil, false
	}
	if strings.HasPrefix(line, "\t") {
		p.recipe(line)
		return nil, false
	}
	text := strings.TrimSpace(stripComment(line))
	if text == "" {
		return nil, false // blank and comment lines keep the rule open
	}
	word := strings.Fields(text)[0]
	switch {
	case conditionals[word]:
		return nil, false
	case word == "define" || strings.HasPrefix(text, "define "):
		p.define, p.current = true, nil
		return nil, false
	case word == "include" || word == "-include" || word == "sinclude":
		p.current = nil
		return strings.Fields(text)[1:], true
	}
	p.current = nil
	p.rule(text)
	return nil, false
}

// recipe records a recipe line for the current targets, without the @, - and
// + prefixes that only change how make runs it.
func (p *parser) recipe(line string) {
	if p.current == nil {
		return
	}
	// Continued lines keep their indentation; collapse it like the shell would.
	cmd := strings.Join(strings.Fields(strings.TrimLeft(strings.TrimSpace(line), "@-+ \t")), " ")
	if cmd == "" || strings.HasPrefix(cmd, "#") {
		return
	}
	for _, t := range p.current {
		p.m.commands[t] = append(p.m.commands[t], cmd)
	}
}

// specialTargets are GNU make's built-in target names, which are not targets
// CI can run.
var specialTargets = map[string]bool{
	".PHONY": true, ".SUFFIXES": true, ".DEFAULT": true, ".PRECIOUS": true, ".INTERMEDIATE": true,
	".NOTINTERMEDIATE": true, ".SECONDARY": true, ".SECONDEXPANSION": true, ".DELETE_ON_ERROR": true,
	".IGNORE": true, ".LOW_RESOLUTION_TIME": true, ".SILENT": true, ".EXPORT_ALL_VARIABLES": true,
	".NOTPARALLEL": true, ".ONESHELL": true, ".POSIX": true, ".WAIT": true,
}

// rule reads a rule line, `targets: prerequisites [; recipe]`, and opens it
// for recipe lines. Variable assignments, target-specific variables, special
// targets, suffix rules, pattern rules and names holding variables define no
// target. A `.SUFFIXES` rule adds to the known suffixes, or clears them when it
// has no prerequisites.
func (p *parser) rule(text string) {
	colon := strings.IndexByte(text, ':')
	eq := strings.IndexByte(text, '=')
	if colon < 0 || (eq >= 0 && eq < colon) {
		return // no rule, or an assignment such as `X = a:b` or `X ?= y`
	}
	rest := strings.TrimLeft(text[colon+1:], ":") // double-colon rules
	if strings.HasPrefix(rest, "=") {
		return // X := y, X ::= y, X :::= y
	}
	prereqs, inline, hasInline := strings.Cut(rest, ";")
	if strings.Contains(prereqs, "=") {
		return // target-specific variable: `t: VAR = x`
	}
	names := strings.Fields(strings.TrimSuffix(strings.TrimSpace(text[:colon]), "&"))
	hasPrereqs := strings.TrimSpace(prereqs) != ""
	if len(names) == 1 && names[0] == ".SUFFIXES" {
		p.suffixes(strings.Fields(prereqs))
	}
	var targets []string
	for _, t := range names {
		if t == "&" || specialTargets[t] || (!hasPrereqs && p.suffixRule(t)) || strings.ContainsAny(t, "%$") {
			continue
		}
		targets = append(targets, t)
		if _, ok := p.m.commands[t]; !ok {
			p.m.commands[t] = nil
		}
		if hasPrereqs {
			p.m.prereqs[t] = true
		}
	}
	p.current = targets
	if hasInline {
		p.recipe("\t" + inline)
	}
}

// suffixes applies a `.SUFFIXES` rule: its prerequisites join the known
// suffixes, and none clears them.
func (p *parser) suffixes(add []string) {
	if len(add) == 0 {
		clear(p.m.suffixes)
	}
	for _, s := range add {
		p.m.suffixes[s] = true
	}
}

// suffixRule reports whether a target is a suffix rule: one known suffix, or two
// joined, such as `.c` or `.c.o`. A suffix rule with prerequisites is an
// ordinary target, so callers check that first.
func (p *parser) suffixRule(t string) bool {
	if p.m.suffixes[t] {
		return true
	}
	for i := 1; i < len(t); i++ {
		if p.m.suffixes[t[:i]] && p.m.suffixes[t[i:]] {
			return true
		}
	}
	return false
}

// stripComment drops a # comment that is not escaped.
func stripComment(line string) string {
	for i := 0; i < len(line); i++ {
		if line[i] == '#' && (i == 0 || line[i-1] != '\\') {
			return line[:i]
		}
	}
	return line
}
