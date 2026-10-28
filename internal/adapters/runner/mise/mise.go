// Package mise reads mise task definitions statically, without running mise:
// tasks from the config files and their includes, and file tasks from the task
// directories.
package mise

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
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
		t := &task{}
		switch v := v.(type) {
		case string, []any:
			t.commands = stringsOf(v)
		case map[string]any:
			t.commands = stringsOf(v["run"])
			t.hidden, _ = v["hide"].(bool)
			for _, alias := range stringsOf(v["alias"]) {
				m.tasks[alias] = &task{hidden: t.hidden}
			}
		default:
			continue
		}
		m.tasks[name] = t
	}
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

// readTaskDir records the executable files under dir as file tasks. A
// `#MISE hide=true` line in a file hides it.
func (m *Mise) readTaskDir(dir string) error {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil || info.Mode()&0o111 == 0 {
			return err
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

// Commands returns a task's run commands; aliases and file tasks have none.
func (m *Mise) Commands(name string) []string {
	if t, ok := m.tasks[name]; ok {
		return t.commands
	}
	return nil
}
