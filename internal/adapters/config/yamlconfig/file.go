package yamlconfig

import "github.com/jimmykurian/ci-parity/internal/domain"

// file mirrors domain.Config with the YAML key names, so the domain type does
// not carry the config file format.
type file struct {
	Runner      string      `yaml:"runner"`
	Taskfile    string      `yaml:"taskfile"`
	Workflows   []string    `yaml:"workflows"`
	Invokers    invokers    `yaml:"invokers"`
	Tools       toolList    `yaml:"tools"`
	Transparent []string    `yaml:"transparent"`
	Allow       []allowRule `yaml:"allow"`
	Baseline    string      `yaml:"baseline"`
}

type invokers struct {
	Shell   []string        `yaml:"shell"`
	Actions []actionInvoker `yaml:"actions"`
}

type actionInvoker struct {
	Uses  string `yaml:"uses"`
	Input string `yaml:"input"`
}

type toolList struct {
	Extend  []string `yaml:"extend"`
	Remove  []string `yaml:"remove"`
	Replace []string `yaml:"replace"`
}

type allowRule struct {
	Match     string   `yaml:"match"`
	Reason    string   `yaml:"reason"`
	Workflows []string `yaml:"workflows"`
}

// fromConfig copies c into a file, so YAML can be decoded over its values.
func fromConfig(c *domain.Config) *file {
	f := &file{
		Runner: c.Runner, Taskfile: c.Taskfile, Workflows: c.Workflows,
		Invokers:    invokers{Shell: c.Invokers.Shell},
		Tools:       toolList(c.Tools),
		Transparent: c.Transparent, Baseline: c.Baseline,
	}
	for _, a := range c.Invokers.Actions {
		f.Invokers.Actions = append(f.Invokers.Actions, actionInvoker(a))
	}
	for _, a := range c.Allow {
		f.Allow = append(f.Allow, allowRule(a))
	}
	return f
}

// config converts f back to the domain type.
func (f *file) config() *domain.Config {
	c := &domain.Config{
		Runner: f.Runner, Taskfile: f.Taskfile, Workflows: f.Workflows,
		Invokers:    domain.Invokers{Shell: f.Invokers.Shell},
		Tools:       domain.ToolList(f.Tools),
		Transparent: f.Transparent, Baseline: f.Baseline,
	}
	for _, a := range f.Invokers.Actions {
		c.Invokers.Actions = append(c.Invokers.Actions, domain.ActionInvoker(a))
	}
	for _, a := range f.Allow {
		c.Allow = append(c.Allow, domain.AllowRule(a))
	}
	return c
}
