// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package mise reads mise task definitions statically, without running mise:
// tasks from the config files and their includes, and file tasks from the task
// directories.
package mise

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// ConfigNames are the config files read from the repository root, lowest
// precedence first, so a later file's task replaces an earlier one's.
var ConfigNames = []string{
	".config/mise/config.toml", ".config/mise.toml", "mise/config.toml",
	".mise/config.toml", "mise.toml", ".mise.toml",
}

// TaskDirs hold file tasks: executables named by their path, `db/migrate` as
// `db:migrate`.
var TaskDirs = []string{".config/mise/tasks", "mise/tasks", ".mise/tasks", "mise-tasks", ".mise-tasks"}

// task is one task or alias.
type task struct {
	hidden   bool
	commands []string
	more     bool   // depends, depends_post, or a run entry that isn't a command
	dir      string // repo-relative directory it runs in, when dirKnown
	dirKnown bool
}

// Mise is a repository's mise tasks, implementing rules.TaskSource.
type Mise struct {
	root  string
	file  string           // repo-relative path of the highest-precedence config file
	tasks map[string]*task // name or alias -> task
}

// Load reads the mise config files and task directories in root.
func Load(root string) (*Mise, error) {
	m := &Mise{root: root, tasks: map[string]*task{}}
	for _, name := range ConfigNames {
		path := filepath.Join(root, filepath.FromSlash(name))
		if !exists(path) {
			continue
		}
		m.file = name
		if err := m.readConfig(path); err != nil {
			return nil, err
		}
	}
	for _, dir := range TaskDirs {
		if err := m.readTaskDir(filepath.Join(root, filepath.FromSlash(dir))); err != nil {
			return nil, err
		}
	}
	if m.file == "" && len(m.tasks) == 0 {
		return nil, fmt.Errorf("no mise config found; looked for %s", strings.Join(ConfigNames, ", "))
	}
	if m.file == "" {
		m.file = "mise.toml"
	}
	return m, nil
}

// exists reports whether path is a file, matching its name exactly against
// the directory listing so case-insensitive file systems agree.
func exists(path string) bool {
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.Name() == filepath.Base(path) && !e.IsDir() {
			return true
		}
	}
	return false
}

// readConfig reads a config file's [tasks] table and task_config.includes.
func (m *Mise) readConfig(path string) error {
	var doc map[string]any
	if _, err := toml.DecodeFile(path, &doc); err != nil {
		return fmt.Errorf("parse %s: %w", m.rel(path), err)
	}
	if tasks, ok := doc["tasks"].(map[string]any); ok {
		m.addTasks(tasks)
	}
	cfg, _ := doc["task_config"].(map[string]any)
	includes, _ := cfg["includes"].([]any)
	for _, inc := range includes {
		name, ok := inc.(string)
		if !ok {
			continue
		}
		if err := m.include(filepath.Join(m.root, filepath.FromSlash(name))); err != nil {
			return err
		}
	}
	return nil
}

// include reads an included task file, whose top-level tables are tasks, or a
// directory of file tasks. A missing include is skipped.
func (m *Mise) include(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if info.IsDir() {
		return m.readTaskDir(path)
	}
	var doc map[string]any
	if _, err := toml.DecodeFile(path, &doc); err != nil {
		return fmt.Errorf("parse %s: %w", m.rel(path), err)
	}
	m.addTasks(doc)
	return nil
}

// addTasks records the tasks of a table: `name = "cmd"`, `name = ["a", "b"]`
// or `[name]` with run, alias and hide.
func (m *Mise) addTasks(tasks map[string]any) {
	for name, v := range tasks {
		t := &task{dirKnown: true} // the project root, unless dir says otherwise
		switch v := v.(type) {
		case string, []any:
			t.commands = stringsOf(v)
		case map[string]any:
			t.commands = stringsOf(v["run"])
			t.more = len(stringsOf(v["depends"])) > 0 || len(stringsOf(v["depends_post"])) > 0 || hasTable(v["run"])
			t.hidden, _ = v["hide"].(bool)
			if d, ok := v["dir"].(string); ok {
				t.dir, t.dirKnown = domain.RepoDir(d)
			}
			for _, alias := range stringsOf(v["alias"]) {
				m.tasks[alias] = &task{hidden: t.hidden}
			}
		default:
			continue
		}
		m.tasks[name] = t
	}
}

// hasTable reports whether a run list holds an entry that isn't a string, such
// as `{ task = "lint" }`.
func hasTable(v any) bool {
	list, _ := v.([]any)
	for _, x := range list {
		if _, ok := x.(string); !ok {
			return true
		}
	}
	return false
}

// stringsOf returns a string, or the strings of a list, with whitespace
// collapsed; anything else gives none.
func stringsOf(v any) []string {
	var out []string
	add := func(s string) {
		if s = strings.Join(strings.Fields(s), " "); s != "" {
			out = append(out, s)
		}
	}
	switch v := v.(type) {
	case string:
		add(v)
	case []any:
		for _, x := range v {
			if s, ok := x.(string); ok {
				add(s)
			}
		}
	}
	return out
}

// readTaskDir records the file tasks under dir: files that are executable or
// start with a `#!` or `#MISE` line. The header test makes the result the same
// on Windows, which has no executable bit. A `#MISE hide=true` line hides a task.
func (m *Mise) readTaskDir(dir string) error {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&0o111 == 0 && !scriptHeader(path) {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := strings.ReplaceAll(filepath.ToSlash(rel), "/", ":")
		m.tasks[name] = &task{hidden: hiddenFile(path)}
		return nil
	})
}

// scriptHeader reports whether the file at path starts with a `#!` shebang or a
// `#MISE` comment, the first lines of a mise file task.
func scriptHeader(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }() // read-only: a close error changes nothing
	head := make([]byte, 8)
	n, _ := io.ReadFull(f, head)
	line := string(head[:n])
	return strings.HasPrefix(line, "#!") || strings.HasPrefix(line, "#MISE") || strings.HasPrefix(line, "# [MISE]")
}

// hiddenFile reports whether a file task's header sets hide=true, in a #MISE or
// # [MISE] comment.
func hiddenFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if l := strings.TrimSpace(line); strings.HasPrefix(l, "#MISE") || strings.HasPrefix(l, "# [MISE]") {
			if strings.Contains(strings.ReplaceAll(l, " ", ""), "hide=true") {
				return true
			}
		}
	}
	return false
}

// rel makes path repo-relative with forward slashes, or leaves it as is when it
// sits outside the repository.
func (m *Mise) rel(path string) string {
	if r, err := filepath.Rel(m.root, path); err == nil && !strings.HasPrefix(r, "..") {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}

// File returns the highest-precedence config file, repo-relative.
func (m *Mise) File() string { return m.file }

// Tasks returns every task and alias, sorted.
func (m *Mise) Tasks() []string {
	names := make([]string, 0, len(m.tasks))
	for name := range m.tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsInternal reports whether a task is hidden.
func (m *Mise) IsInternal(name string) bool {
	t, ok := m.tasks[name]
	return ok && t.hidden
}

// Dir returns the repo-relative directory a task runs in: the project root, or
// its dir, relative to the root. A templated dir, such as {{cwd}}, is unknown.
func (m *Mise) Dir(name string) (string, bool) {
	if t, ok := m.tasks[name]; ok {
		return t.dir, t.dirKnown
	}
	return "", false
}

// Reaches reports true: mise reads the config files of every parent directory,
// so the project's tasks run from anywhere in the repository.
func (m *Mise) Reaches(string) bool { return true }

// RunsMore reports whether a task has depends or depends_post, or runs another
// task from its run list.
func (m *Mise) RunsMore(name string) bool {
	t, ok := m.tasks[name]
	return ok && t.more
}

// Commands returns a task's run commands; aliases and file tasks have none.
func (m *Mise) Commands(name string) []string {
	if t, ok := m.tasks[name]; ok {
		return t.commands
	}
	return nil
}
