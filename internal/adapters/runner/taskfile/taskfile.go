// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package taskfile reads go-task v3 Taskfiles statically, without running task.
package taskfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// DefaultNames are the file names searched, in order, when no Taskfile path is
// configured and when an include points at a directory.
var DefaultNames = []string{
	"Taskfile.yml", "taskfile.yml", "Taskfile.yaml", "taskfile.yaml",
	"Taskfile.dist.yml", "taskfile.dist.yml", "Taskfile.dist.yaml", "taskfile.dist.yaml",
}

// maxDepth caps include recursion.
const maxDepth = 10

// Taskfile is a resolved Taskfile with its includes flattened into
// namespaced task names. It implements runner.Runner.
type Taskfile struct {
	root     string
	file     string
	internal map[string]bool     // every invocable name -> internal
	commands map[string][]string // canonical task name -> command strings
	more     map[string]bool     // canonical task name -> does more than run its commands
	dirs     map[string]string   // canonical task name -> repo-relative directory it runs in
}

// Load reads the Taskfile at path, relative to root, or discovers one of
// DefaultNames in root when path is empty, and resolves its includes.
func Load(root, path string) (*Taskfile, error) {
	full := filepath.Join(root, filepath.FromSlash(path))
	if path == "" {
		var ok bool
		if full, ok = discover(root); !ok {
			return nil, fmt.Errorf("no Taskfile found; looked for %s", strings.Join(DefaultNames[:4], ", "))
		}
	} else if filepath.IsAbs(path) {
		full = path
	}
	t := &Taskfile{root: root, internal: map[string]bool{}, commands: map[string][]string{}, more: map[string]bool{}, dirs: map[string]string{}}
	t.file = t.rel(full)
	if _, err := os.Stat(full); err != nil {
		return nil, fmt.Errorf("read Taskfile %s: %w", t.file, err)
	}
	if err := t.load(full, []string{""}, false, filepath.Dir(full), nil); err != nil {
		return nil, err
	}
	return t, nil
}

// load adds the tasks of the Taskfile at path under each of prefixes, then
// follows its includes. dir is where its tasks run by default: the root
// Taskfile's directory, or an include's dir:; "" when it isn't known. stack
// holds the files being loaded, to stop cycles.
func (t *Taskfile) load(path string, prefixes []string, internal bool, dir string, stack []string) error {
	if len(stack) > maxDepth {
		return nil
	}
	for _, s := range stack {
		if s == path {
			return nil
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", t.rel(path), err)
	}
	f, err := parse(data)
	if err != nil {
		return fmt.Errorf("parse %s: %w", t.rel(path), err)
	}
	t.addTasks(f.tasks, prefixes, internal, dir)
	stack = append(stack, path)
	for _, inc := range f.includes {
		target, err := resolve(filepath.Dir(path), inc.path)
		if err != nil {
			if inc.optional {
				continue
			}
			if errors.Is(err, fs.ErrNotExist) {
				err = fmt.Errorf("%s does not exist", t.rel(target))
			}
			return fmt.Errorf("include %q in %s: %w", inc.namespace, t.rel(path), err)
		}
		if err := t.load(target, inc.prefixes(prefixes), internal || inc.internal, inc.runDir(filepath.Dir(path), dir), stack); err != nil {
			return err
		}
	}
	return nil
}

// addTasks registers tasks under every prefix. Commands and directories are
// kept only under the first prefix, the canonical namespace. A default task
// also makes the bare namespace invocable, as `task <namespace>` runs it.
func (t *Taskfile) addTasks(tasks []task, prefixes []string, internal bool, dir string) {
	for _, tk := range tasks {
		hidden := internal || tk.internal
		for i, p := range prefixes {
			t.add(p+tk.name, hidden)
			if i == 0 && len(tk.cmds) > 0 {
				t.commands[p+tk.name] = tk.cmds
				t.more[p+tk.name] = tk.more
				if d, ok := t.taskDir(dir, tk.dir); ok {
					t.dirs[p+tk.name] = d
				}
			}
			for _, a := range tk.aliases {
				t.add(p+a, hidden)
			}
			if tk.name == "default" && p != "" {
				t.add(strings.TrimSuffix(p, ":"), hidden)
			}
		}
	}
}

// add records a task name. The first definition wins, so a task stays internal
// or not as first seen.
func (t *Taskfile) add(name string, internal bool) {
	if _, ok := t.internal[name]; !ok {
		t.internal[name] = internal
	}
}

// runDir returns where an include's tasks run: its dir:, relative to the
// including Taskfile's directory from, or else parent, where the including
// file's tasks run. It is "" when a template makes it unknown.
func (inc include) runDir(from, parent string) string {
	switch {
	case inc.dir == "":
		return parent
	case strings.Contains(inc.dir, "{{"):
		return ""
	case filepath.IsAbs(inc.dir):
		return inc.dir
	}
	return filepath.Join(from, filepath.FromSlash(inc.dir))
}

// taskDir returns the repo-relative directory a task runs in: its dir:,
// relative to base, or base itself. It returns false when base is unknown, the
// dir: is a template, or the directory is outside the repository.
func (t *Taskfile) taskDir(base, dir string) (string, bool) {
	if base == "" || strings.Contains(dir, "{{") {
		return "", false
	}
	if dir != "" {
		if filepath.IsAbs(dir) {
			base = dir
		} else {
			base = filepath.Join(base, filepath.FromSlash(dir))
		}
	}
	rel, err := filepath.Rel(t.root, base)
	if err != nil {
		return "", false
	}
	return domain.RepoDir(filepath.ToSlash(rel))
}

// prefixes returns the name prefixes for an include's tasks under each of
// parents: the namespace and its aliases, or the parent itself when flattened.
func (inc include) prefixes(parents []string) []string {
	var out []string
	for _, p := range parents {
		if inc.flatten {
			out = append(out, p)
			continue
		}
		for _, ns := range append([]string{inc.namespace}, inc.aliases...) {
			out = append(out, p+ns+":")
		}
	}
	return out
}

// resolve returns the Taskfile an include path names, relative to dir. On a
// missing file it also returns the path it looked for.
func resolve(dir, path string) (string, error) {
	if strings.Contains(path, "{{") {
		return "", fmt.Errorf("path %q uses a template, which cannot be read statically", path)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, filepath.FromSlash(path))
	}
	info, err := os.Stat(path)
	if err != nil {
		return path, fs.ErrNotExist
	}
	if !info.IsDir() {
		return path, nil
	}
	if found, ok := discover(path); ok {
		return found, nil
	}
	return filepath.Join(path, DefaultNames[0]), fs.ErrNotExist
}

// discover returns the first of DefaultNames present in dir. Names are matched
// exactly against the directory listing, so case-insensitive file systems
// report the same file as case-sensitive ones.
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
	for _, name := range DefaultNames {
		if present[name] {
			return filepath.Join(dir, name), true
		}
	}
	return "", false
}

// rel makes path repo-relative with forward slashes, or leaves it as is when it
// sits outside the repository.
func (t *Taskfile) rel(path string) string {
	if r, err := filepath.Rel(t.root, path); err == nil && !strings.HasPrefix(r, "..") {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}

// File returns the root Taskfile path, repo-relative.
func (t *Taskfile) File() string { return t.file }

// Tasks returns every invocable name, including aliases, sorted.
func (t *Taskfile) Tasks() []string {
	names := make([]string, 0, len(t.internal))
	for n := range t.internal {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Matches reports false: every name is listed by Tasks.
func (*Taskfile) Matches(string) bool { return false }

// IsInternal reports whether the task called name is internal.
func (t *Taskfile) IsInternal(name string) bool { return t.internal[name] }

// Commands returns the command strings of the task defined under name.
func (t *Taskfile) Commands(name string) []string { return t.commands[name] }

// RunsMore reports whether the task defined under name has deps:, or calls
// another task, loops or defers in its cmds:.
func (t *Taskfile) RunsMore(name string) bool { return t.more[name] }

// Reaches reports whether `task` called from dir, repo-relative, reads the root
// Taskfile: dir is at or below the Taskfile's directory, and neither dir nor a
// directory between them holds a Taskfile, which task would find first.
func (t *Taskfile) Reaches(dir string) bool {
	return domain.Reaches(filepath.Join(t.root, filepath.FromSlash(t.file)), filepath.Join(t.root, filepath.FromSlash(dir)), discover)
}

// Dir returns the repo-relative directory the task defined under name runs in,
// or false when it can't be known statically.
func (t *Taskfile) Dir(name string) (string, bool) {
	d, ok := t.dirs[name]
	return d, ok
}
