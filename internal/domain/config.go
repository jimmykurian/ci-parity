package domain

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
