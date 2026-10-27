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
	t := &Taskfile{root: root, internal: map[string]bool{}, commands: map[string][]string{}}
	t.file = t.rel(full)
	if _, err := os.Stat(full); err != nil {
		return nil, fmt.Errorf("read Taskfile %s: %w", t.file, err)
	}
	if err := t.load(full, []string{""}, false, nil); err != nil {
		return nil, err
	}
	return t, nil
}

// load adds the tasks of the Taskfile at path under each of prefixes, then
// follows its includes. stack holds the files being loaded, to stop cycles.
func (t *Taskfile) load(path string, prefixes []string, internal bool, stack []string) error {
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
	t.addTasks(f.tasks, prefixes, internal)
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
		if err := t.load(target, inc.prefixes(prefixes), internal || inc.internal, stack); err != nil {
			return err
		}
	}
	return nil
}

// addTasks registers tasks under every prefix. Commands are kept only under
// the first prefix, the canonical namespace. A default task also makes the
// bare namespace invocable, as `task <namespace>` runs it.
func (t *Taskfile) addTasks(tasks []task, prefixes []string, internal bool) {
	for _, tk := range tasks {
		hidden := internal || tk.internal
		for i, p := range prefixes {
			t.add(p+tk.name, hidden)
			if i == 0 && len(tk.cmds) > 0 {
				t.commands[p+tk.name] = tk.cmds
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

func (t *Taskfile) add(name string, internal bool) {
	if _, ok := t.internal[name]; !ok {
		t.internal[name] = internal
	}
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

// IsInternal reports whether the task called name is internal.
func (t *Taskfile) IsInternal(name string) bool { return t.internal[name] }

// Commands returns the command strings of the task defined under name.
func (t *Taskfile) Commands(name string) []string { return t.commands[name] }
