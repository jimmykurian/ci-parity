// Package config loads, defaults and validates .ci-parity.yml.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultPath is the config file name, relative to the repo root.
const DefaultPath = ".ci-parity.yml"

// Config is the resolved ci-parity configuration. See docs/SPEC.md section 4.2.
type Config struct {
	Runner      string      `yaml:"runner"`
	Taskfile    string      `yaml:"taskfile"`
	Workflows   []string    `yaml:"workflows"`
	Invokers    Invokers    `yaml:"invokers"`
	Tools       ToolList    `yaml:"tools"`
	Transparent []string    `yaml:"transparent"`
	Allow       []AllowRule `yaml:"allow"`
	Baseline    string      `yaml:"baseline"`
}

// Invokers describes how CI calls the task runner.
type Invokers struct {
	Shell   []string        `yaml:"shell"`
	Actions []ActionInvoker `yaml:"actions"`
}

// ActionInvoker is a local action whose input names the task to run.
type ActionInvoker struct {
	Uses  string `yaml:"uses"`
	Input string `yaml:"input"`
}

// ToolList adjusts the built-in tool list.
type ToolList struct {
	Extend  []string `yaml:"extend"`
	Remove  []string `yaml:"remove"`
	Replace []string `yaml:"replace"`
}

// AllowRule allows inline commands that match a command glob.
type AllowRule struct {
	Match     string   `yaml:"match"`
	Reason    string   `yaml:"reason"`
	Workflows []string `yaml:"workflows"`
}

// BuiltinTools is the default tools list, matched on the command's basename.
var BuiltinTools = []string{
	// node
	"npm", "npx", "pnpm", "pnpx", "yarn", "bun", "bunx", "node", "deno",
	// python
	"python", "python3", "pip", "pip3", "pipx", "uv", "uvx", "poetry", "pytest", "ruff", "mypy", "tox", "nox",
	// dotnet, go, rust, jvm
	"dotnet", "go", "golangci-lint", "cargo", "mvn", "mvnw", "gradle", "gradlew",
	// infra
	"terraform", "tofu", "docker", "docker-compose", "kubectl", "helm", "az", "aws", "gcloud", "pulumi",
	// quality
	"eslint", "prettier", "biome", "vitest", "jest", "playwright", "cypress", "k6", "shellcheck", "actionlint",
	// other task runners, which bypass the configured one
	"make", "just", "mise", "cargo-make",
}

// Default returns the configuration used when no config file exists.
func Default() *Config {
	return &Config{
		Runner: "taskfile",
		Workflows: []string{
			".github/workflows/*.yml",
			".github/workflows/*.yaml",
			".github/actions/**/action.yml",
			".github/actions/**/action.yaml",
		},
		Invokers: Invokers{
			Shell:   []string{"task", "go-task", "go tool task"},
			Actions: []ActionInvoker{{Uses: "./.github/actions/run-task", Input: "task"}},
		},
		Transparent: []string{"sudo", "env", "time", "nice", "ionice", "timeout", "nohup", "xvfb-run", "chronic"},
		Baseline:    ".ci-parity-baseline.json",
	}
}

// Load reads the config at path, relative to root unless absolute. A missing
// file yields the defaults unless required is true.
func Load(root, path string, required bool) (*Config, error) {
	full := path
	if !filepath.IsAbs(full) {
		full = filepath.Join(root, path)
	}
	data, err := os.ReadFile(full)
	if errors.Is(err, fs.ErrNotExist) && !required {
		return Default(), nil
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
func Parse(data []byte) (*Config, error) {
	cfg := Default()
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
	if err := root.Decode(cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Runner != "taskfile" {
		return fmt.Errorf("runner: unsupported runner %q; v0.1 supports taskfile", c.Runner)
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

// ToolSet returns the effective tools list as a set.
func (c *Config) ToolSet() map[string]bool {
	base := BuiltinTools
	if len(c.Tools.Replace) > 0 {
		base = c.Tools.Replace
	}
	set := make(map[string]bool, len(base)+len(c.Tools.Extend))
	for _, t := range base {
		set[t] = true
	}
	for _, t := range c.Tools.Extend {
		set[t] = true
	}
	for _, t := range c.Tools.Remove {
		delete(set, t)
	}
	return set
}
