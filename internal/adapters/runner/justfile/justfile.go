// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package justfile reads justfiles statically, without running just: recipes
// with their parameters, aliases, imports and modules, and the body lines used
// to suggest a recipe for an inline command.
package justfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Names are the file names just looks for in a directory, in its order.
var Names = []string{"justfile", "Justfile", ".justfile"}

// maxDepth caps import and module recursion.
const maxDepth = 10

// recipe is one recipe or alias, under its full name (`mod::name`).
type recipe struct {
	min, max int      // how many arguments it takes; max is -1 when variadic
	private  bool     // [private], or a name starting with _
	commands []string // body lines, for suggestions
	alias    string   // for an alias, the recipe it names
}

// Justfile is a justfile with its imports and modules read, implementing
// rules.TaskSource.
type Justfile struct {
	root    string
	file    string             // repo-relative path of the root justfile
	recipes map[string]*recipe // full name -> recipe
	modules map[string]bool    // module paths, such as "db" and "db::seed"
}

// Load discovers the justfile in root and reads it with its imports and modules.
func Load(root string) (*Justfile, error) {
	path, ok := discover(root)
	if !ok {
		return nil, fmt.Errorf("no justfile found; looked for %s", strings.Join(Names, ", "))
	}
	j := &Justfile{root: root, file: filepath.Base(path), recipes: map[string]*recipe{}, modules: map[string]bool{}}
	if err := j.read(path, "", 0, nil); err != nil {
		return nil, err
	}
	j.resolveAliases()
	return j, nil
}

// discover returns the first of Names in dir, matching names exactly.
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

// read parses the justfile at path into the module prefix (`db::`, or "" at
// the root). stack holds the files being read, so a cycle stops.
func (j *Justfile) read(path, prefix string, depth int, stack []string) error {
	if depth > maxDepth {
		return nil
	}
	for _, f := range stack {
		if f == path {
			return nil
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", j.rel(path), err)
	}
	p := parser{j: j, prefix: prefix}
	p.parse(string(data))
	dir, stack := filepath.Dir(path), append(stack, path)
	for _, imp := range p.imports {
		target := filepath.Join(dir, filepath.FromSlash(imp.path))
		if _, err := os.Stat(target); err != nil {
			if imp.optional {
				continue
			}
			return fmt.Errorf("%s: import %q not found", j.rel(path), imp.path)
		}
		if err := j.read(target, prefix, depth+1, stack); err != nil {
			return err
		}
	}
	for _, mod := range p.modules {
		target, ok := moduleFile(dir, mod)
		if !ok {
			if mod.optional {
				continue
			}
			return fmt.Errorf("%s: module %q not found", j.rel(path), mod.name)
		}
		j.modules[prefix+mod.name] = true
		if err := j.read(target, prefix+mod.name+"::", depth+1, stack); err != nil {
			return err
		}
	}
	return nil
}

// moduleFile finds a module's file: the given path, or name.just,
// name/mod.just, name/justfile or name/.justfile beside the declaring file.
func moduleFile(dir string, mod module) (string, bool) {
	candidates := []string{mod.name + ".just", filepath.Join(mod.name, "mod.just"), filepath.Join(mod.name, "justfile"), filepath.Join(mod.name, ".justfile")}
	if mod.path != "" {
		candidates = []string{filepath.FromSlash(mod.path)}
	}
	for _, c := range candidates {
		path := filepath.Join(dir, c)
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				if found, ok := discover(path); ok {
					return found, true
				}
				continue
			}
			return path, true
		}
	}
	return "", false
}

// resolveAliases gives each alias its recipe's parameters, so a call through
// an alias consumes the same arguments.
func (j *Justfile) resolveAliases() {
	for _, r := range j.recipes {
		if r.alias == "" {
			continue
		}
		if target, ok := j.recipes[r.alias]; ok && target.alias == "" {
			r.min, r.max = target.min, target.max
		}
	}
}

// rel makes path repo-relative with forward slashes, or leaves it as is when it
// sits outside the repository.
func (j *Justfile) rel(path string) string {
	if r, err := filepath.Rel(j.root, path); err == nil && !strings.HasPrefix(r, "..") {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}

// File returns the root justfile's name, repo-relative.
func (j *Justfile) File() string { return j.file }

// Tasks returns every recipe and alias, sorted.
func (j *Justfile) Tasks() []string {
	names := make([]string, 0, len(j.recipes))
	for name := range j.recipes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsInternal reports whether a recipe or alias is private.
func (j *Justfile) IsInternal(name string) bool {
	r, ok := j.recipes[name]
	return ok && r.private
}

// Commands returns a recipe's body lines; an alias has none, so suggestions
// name the recipe itself.
func (j *Justfile) Commands(name string) []string {
	if r, ok := j.recipes[name]; ok {
		return r.commands
	}
	return nil
}
