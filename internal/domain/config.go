// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

// Config is the resolved ci-parity configuration. See docs/reference/configuration.md.
type Config struct {
	Runner      string
	Taskfile    string
	Workflows   []string
	Invokers    Invokers
	Tools       ToolList
	Transparent []string
	Allow       []AllowRule
	// BuiltinAllow turns on the built-in allow set (see BuiltinAllow).
	BuiltinAllow bool
	Baseline     string
}

// Invokers describes how CI calls the task runner.
type Invokers struct {
	Shell   []string
	Actions []ActionInvoker
}

// ActionInvoker is a local action whose input names the task to run.
type ActionInvoker struct {
	Uses  string
	Input string
}

// ToolList adjusts the built-in tool list.
type ToolList struct {
	Extend  []string
	Remove  []string
	Replace []string
}

// AllowRule allows inline commands that match a command glob.
type AllowRule struct {
	Match     string
	Reason    string
	Workflows []string
}

// BuiltinTools is the default tools list, matched on the command's basename.
var BuiltinTools = []string{
	// node
	"npm", "npx", "pnpm", "pnpx", "pnx", "yarn", "bun", "bunx", "node", "deno",
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
			Shell:   DefaultShellInvokers(RunnerTaskfile),
			Actions: []ActionInvoker{{Uses: "./.github/actions/run-task", Input: "task"}},
		},
		Transparent:  []string{"sudo", "env", "time", "nice", "ionice", "timeout", "nohup", "xvfb-run", "chronic"},
		BuiltinAllow: true,
		Baseline:     ".ci-parity-baseline.json",
	}
}

// Runner names, the values of the config's runner key.
const (
	RunnerTaskfile = "taskfile"
	RunnerMake     = "make"
	RunnerJust     = "just"
	RunnerMise     = "mise"
)

// Runners lists the runners with a reader.
var Runners = []string{RunnerTaskfile, RunnerMake, RunnerJust, RunnerMise}

// runnerTools are each runner's own commands. The configured runner's leave
// the built-in tools list; every other runner's are bypasses.
var runnerTools = map[string][]string{
	RunnerTaskfile: {"task", "go-task"},
	RunnerMake:     {"make"},
	RunnerJust:     {"just"},
	RunnerMise:     {"mise"},
}

// TaskModule is task's Go package, which `go tool` and `go run` can run.
const TaskModule = "github.com/go-task/task/v3/cmd/task"

// DefaultShellInvokers returns the shell invokers a runner uses when the
// config sets none.
func DefaultShellInvokers(runner string) []string {
	switch runner {
	case RunnerMake:
		return []string{"make", "gmake"}
	case RunnerJust:
		return []string{"just"}
	case RunnerMise:
		return []string{"mise run", "mise r", "mise tasks run"}
	default:
		return []string{"task", "go-task", "go tool task", "go tool " + TaskModule, "go run " + TaskModule}
	}
}

// ToolSet returns the effective tools list as a set. Unless the list is
// replaced, the configured runner's own commands leave it and the other
// runners' join it as bypasses.
func (c *Config) ToolSet() map[string]bool {
	base := BuiltinTools
	if len(c.Tools.Replace) > 0 {
		base = c.Tools.Replace
	}
	set := make(map[string]bool, len(base)+len(c.Tools.Extend))
	for _, t := range base {
		set[t] = true
	}
	if len(c.Tools.Replace) == 0 {
		for runner, tools := range runnerTools {
			for _, t := range tools {
				set[t] = runner != c.Runner
			}
		}
		for t, on := range set {
			if !on {
				delete(set, t)
			}
		}
	}
	for _, t := range c.Tools.Extend {
		set[t] = true
	}
	for _, t := range c.Tools.Remove {
		delete(set, t)
	}
	return set
}
