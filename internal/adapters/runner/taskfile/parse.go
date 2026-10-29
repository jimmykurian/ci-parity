// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package taskfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// errVersion is returned for a Taskfile version other than 3.
var errVersion = errors.New("unsupported Taskfile version")

// file is the part of one Taskfile that ci-parity reads.
type file struct {
	tasks    []task
	includes []include
}

// task is one task: its name, aliases, whether it is internal, its commands,
// whether it does more than run them, and its directory.
type task struct {
	name     string
	aliases  []string
	internal bool
	cmds     []string
	more     bool   // deps:, or a task:, for: or defer: entry in cmds:
	dir      string // dir:, as written
}

// include is one `includes` entry: its namespace and aliases, the file it reads,
// whether it is optional, internal or flattened, and the directory its tasks run in.
type include struct {
	namespace string
	path      string
	aliases   []string
	optional  bool
	internal  bool
	flatten   bool
	dir       string // dir:, as written
}

// parse decodes one Taskfile. Keys ci-parity does not need are ignored.
func parse(data []byte) (*file, error) {
	var doc yaml.Node
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&doc); err != nil {
		if errors.Is(err, io.EOF) {
			return &file{}, nil
		}
		return nil, err
	}
	f := &file{}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return f, nil
	}
	top := doc.Content[0]
	if v := scalar(get(top, "version")); v != "" && v != "3" && !strings.HasPrefix(v, "3.") {
		return nil, fmt.Errorf("%w %q; ci-parity reads version 3", errVersion, v)
	}
	each(get(top, "tasks"), func(name string, n *yaml.Node) {
		f.tasks = append(f.tasks, parseTask(name, n))
	})
	each(get(top, "includes"), func(ns string, n *yaml.Node) {
		f.includes = append(f.includes, parseInclude(ns, n))
	})
	return f, nil
}

// parseTask reads a task in its full form or its string and list shorthands.
func parseTask(name string, n *yaml.Node) task {
	t := task{name: name}
	switch n.Kind {
	case yaml.ScalarNode:
		t.cmds = []string{n.Value}
	case yaml.SequenceNode:
		t.cmds, t.more = commands(n)
	case yaml.MappingNode:
		t.aliases = scalars(get(n, "aliases"))
		t.internal = scalar(get(n, "internal")) == "true"
		t.dir = scalar(get(n, "dir"))
		if c := scalar(get(n, "cmd")); c != "" {
			t.cmds = append(t.cmds, c)
		}
		cmds, more := commands(get(n, "cmds"))
		t.cmds = append(t.cmds, cmds...)
		deps := get(n, "deps")
		t.more = more || (deps != nil && len(deps.Content) > 0)
	}
	return t
}

// commands returns the string entries and cmd: entries of a cmds list,
// skipping task:, for: and defer: entries, and whether it skipped any.
func commands(n *yaml.Node) ([]string, bool) {
	if n == nil || n.Kind != yaml.SequenceNode {
		return nil, false
	}
	var out []string
	more := false
	for _, c := range n.Content {
		switch {
		case c.Kind == yaml.ScalarNode:
			out = append(out, c.Value)
		case c.Kind == yaml.MappingNode && get(c, "task") == nil && get(c, "for") == nil && get(c, "defer") == nil:
			if v := scalar(get(c, "cmd")); v != "" {
				out = append(out, v)
			}
		case c.Kind == yaml.MappingNode:
			more = true
		}
	}
	return out, more
}

// parseInclude reads an include given as a path or as a mapping.
func parseInclude(ns string, n *yaml.Node) include {
	inc := include{namespace: ns}
	if n.Kind == yaml.ScalarNode {
		inc.path = n.Value
		return inc
	}
	inc.path = scalar(get(n, "taskfile"))
	inc.aliases = scalars(get(n, "aliases"))
	inc.optional = scalar(get(n, "optional")) == "true"
	inc.internal = scalar(get(n, "internal")) == "true"
	inc.flatten = scalar(get(n, "flatten")) == "true"
	inc.dir = scalar(get(n, "dir"))
	return inc
}

// each calls fn for every entry of mapping n, in file order.
func each(n *yaml.Node, fn func(key string, value *yaml.Node)) {
	if n == nil || n.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		fn(n.Content[i].Value, n.Content[i+1])
	}
}

// get returns the value of key in mapping n, or nil when n isn't a mapping or has
// no such key.
func get(n *yaml.Node, key string) *yaml.Node {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}

// scalar returns n's value when it is a scalar, or "".
func scalar(n *yaml.Node) string {
	if n == nil || n.Kind != yaml.ScalarNode {
		return ""
	}
	return n.Value
}

// scalars returns the scalar items of sequence n, skipping any others.
func scalars(n *yaml.Node) []string {
	if n == nil || n.Kind != yaml.SequenceNode {
		return nil
	}
	var out []string
	for _, c := range n.Content {
		if c.Kind == yaml.ScalarNode {
			out = append(out, c.Value)
		}
	}
	return out
}
