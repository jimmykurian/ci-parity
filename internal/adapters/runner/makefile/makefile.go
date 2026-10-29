// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package makefile reads Makefiles statically, without running make: their
// targets, and the recipe lines used to suggest a target for an inline command.
package makefile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Names are the file names make looks for, in its order.
var Names = []string{"GNUmakefile", "makefile", "Makefile"}

// maxDepth caps include recursion.
const maxDepth = 10

// Makefile is a Makefile with its includes read, implementing rules.TaskSource.
type Makefile struct {
	root     string
	file     string              // repo-relative path of the root Makefile
	commands map[string][]string // target -> recipe lines
	prereqs  map[string]bool     // targets with prerequisites
	suffixes map[string]bool     // the known suffixes, for suffix rules
}

// defaultSuffixes is GNU make's default suffix list.
var defaultSuffixes = []string{
	".out", ".a", ".ln", ".o", ".c", ".cc", ".C", ".cpp", ".p", ".f", ".F", ".m", ".r", ".y",
	".l", ".ym", ".lm", ".s", ".S", ".mod", ".sym", ".def", ".h", ".info", ".dvi", ".tex",
	".texinfo", ".texi", ".txinfo", ".w", ".ch", ".web", ".sh", ".elc", ".el",
}

// Load discovers the Makefile in root and reads it and its includes.
func Load(root string) (*Makefile, error) {
	path, ok := discover(root)
	if !ok {
		return nil, fmt.Errorf("no Makefile found; looked for %s", strings.Join(Names, ", "))
	}
	m := &Makefile{root: root, file: filepath.Base(path), commands: map[string][]string{}, prereqs: map[string]bool{}, suffixes: map[string]bool{}}
	for _, s := range defaultSuffixes {
		m.suffixes[s] = true
	}
	if err := m.read(path, 0, nil); err != nil {
		return nil, err
	}
	return m, nil
}

// discover returns the first of Names in dir, matching names exactly so that
// case-insensitive file systems agree with case-sensitive ones.
func discover(dir string) (string, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	present := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			present[e.Name()] = true
		}
	}
	for _, name := range Names {
		if present[name] {
			return filepath.Join(dir, name), true
		}
	}
	return "", false
}

// read parses the Makefile at path. stack holds the files being read, so an
// include cycle stops instead of recursing.
func (m *Makefile) read(path string, depth int, stack []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", m.rel(path), err)
	}
	p := parser{m: m}
	for _, line := range logicalLines(string(data)) {
		if inc, ok := p.line(line); ok {
			m.include(inc, depth, append(stack, path))
		}
	}
	return nil
}

// include reads the files an include directive names. make resolves them from
// its working directory, the repository root here. A file that is missing,
// whose name holds a variable, or that is already being read is skipped: make
// may generate it, and its targets cannot be known statically.
func (m *Makefile) include(names []string, depth int, stack []string) {
	if depth+1 > maxDepth {
		return
	}
	for _, name := range names {
		if strings.Contains(name, "$") {
			continue
		}
		path := filepath.Join(m.root, filepath.FromSlash(name))
		if filepath.IsAbs(name) {
			path = name
		}
		if contains(stack, path) {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}
		_ = m.read(path, depth+1, stack)
	}
}

// contains reports whether s is in list.
func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// rel makes path repo-relative with forward slashes, or leaves it as is when it
// sits outside the repository.
func (m *Makefile) rel(path string) string {
	if r, err := filepath.Rel(m.root, path); err == nil && !strings.HasPrefix(r, "..") {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}

// File returns the root Makefile's name, repo-relative.
func (m *Makefile) File() string { return m.file }

// Tasks returns every target, sorted.
func (m *Makefile) Tasks() []string {
	names := make([]string, 0, len(m.commands))
	for name := range m.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsInternal reports false: make has no internal targets.
func (m *Makefile) IsInternal(string) bool { return false }

// RunsMore reports whether a target has prerequisites, which make brings up to
// date first.
func (m *Makefile) RunsMore(name string) bool { return m.prereqs[name] }

// Commands returns the recipe lines of a target.
func (m *Makefile) Commands(name string) []string { return m.commands[name] }

// Dir returns the repository root: make reads the Makefile there, and its
// recipes run in the directory make starts in.
func (m *Makefile) Dir(string) (string, bool) { return "", true }

// Reaches reports whether dir is the root: make reads only the Makefile in the
// directory it starts in.
func (m *Makefile) Reaches(dir string) bool { return dir == "" }
