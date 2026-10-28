// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package yamlconfig loads, defaults and validates .ci-parity.yml.
package yamlconfig

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// DefaultPath is the config file name, relative to the repo root.
const DefaultPath = ".ci-parity.yml"

// Load reads the config at path, relative to root unless absolute. A missing
// file yields the defaults unless required is true.
func Load(root, path string, required bool) (*domain.Config, error) {
	full := path
	if !filepath.IsAbs(full) {
		full = filepath.Join(root, path)
	}
	data, err := os.ReadFile(full)
	if errors.Is(err, fs.ErrNotExist) && !required {
		return domain.Default(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("config %s: %w", filepath.ToSlash(path), err)
	}
	return cfg, nil
}

// Parse decodes config YAML over the defaults and validates it.
func Parse(data []byte) (*domain.Config, error) {
	cfg := domain.Default()
	var doc yaml.Node
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&doc); err != nil {
		if errors.Is(err, io.EOF) {
			return cfg, nil
		}
		return nil, err
	}
	if len(doc.Content) == 0 || isNull(doc.Content[0]) {
		return cfg, nil
	}
	root := doc.Content[0]
	if err := fileSchema.check(root, ""); err != nil {
		return nil, err
	}
	f := fromConfig(cfg)
	if err := root.Decode(f); err != nil {
		return nil, err
	}
	cfg = f.config()
	// Invokers default by runner; keys the file sets are kept as written.
	if cfg.Runner != domain.RunnerTaskfile {
		if !has(root, "invokers", "shell") {
			cfg.Invokers.Shell = domain.DefaultShellInvokers(cfg.Runner)
		}
		if !has(root, "invokers", "actions") {
			cfg.Invokers.Actions = nil
		}
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate checks the rules that span keys or values, which decoding alone can't:
// a known runner, taskfile only with runner taskfile, tools.replace on its own,
// and no empty invoker or allow entry.
func validate(c *domain.Config) error {
	if !slices.Contains(domain.Runners, c.Runner) {
		return fmt.Errorf("runner: unsupported runner %q; use %s", c.Runner, strings.Join(domain.Runners, ", "))
	}
	if c.Taskfile != "" && c.Runner != domain.RunnerTaskfile {
		return fmt.Errorf("taskfile: applies only to runner taskfile, not %s", c.Runner)
	}
	if len(c.Tools.Replace) > 0 && (len(c.Tools.Extend) > 0 || len(c.Tools.Remove) > 0) {
		return errors.New("tools.replace cannot be combined with tools.extend or tools.remove")
	}
	for i, inv := range c.Invokers.Shell {
		if inv == "" {
			return fmt.Errorf("invokers.shell[%d] is empty", i)
		}
	}
	for i, a := range c.Invokers.Actions {
		if a.Uses == "" || a.Input == "" {
			return fmt.Errorf("invokers.actions[%d] needs both uses and input", i)
		}
	}
	for i, a := range c.Allow {
		if a.Match == "" {
			return fmt.Errorf("allow[%d].match is required", i)
		}
		if a.Reason == "" {
			return fmt.Errorf("allow[%d].reason is required", i)
		}
	}
	return nil
}

// has reports whether the mapping n holds the nested keys, even with a null value.
func has(n *yaml.Node, keys ...string) bool {
	for _, k := range keys {
		if n == nil || n.Kind != yaml.MappingNode {
			return false
		}
		var next *yaml.Node
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Value == k {
				next = n.Content[i+1]
			}
		}
		if next == nil {
			return false
		}
		n = next
	}
	return true
}
