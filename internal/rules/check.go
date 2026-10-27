package rules

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/config"
	"github.com/jimmykurian/ci-parity/internal/shell"
	"github.com/jimmykurian/ci-parity/internal/workflow"
)

// Check evaluates every step of files against cfg and returns sorted findings.
func Check(cfg *config.Config, files []*workflow.File) []Finding {
	c := checker{cfg: cfg, opts: shellOptions(cfg)}
	for _, f := range files {
		for i := range f.Steps {
			c.step(f, &f.Steps[i])
		}
	}
	Sort(c.findings)
	return c.findings
}

func shellOptions(cfg *config.Config) shell.Options {
	opts := shell.Options{Tools: cfg.ToolSet(), Transparent: map[string]bool{}}
	for _, t := range cfg.Transparent {
		opts.Transparent[t] = true
	}
	for _, inv := range cfg.Invokers.Shell {
		opts.Invokers = append(opts.Invokers, strings.Fields(inv))
	}
	return opts
}

type checker struct {
	cfg      *config.Config
	opts     shell.Options
	findings []Finding
}

func (c *checker) step(f *workflow.File, s *workflow.Step) {
	if s.Run == nil {
		return
	}
	switch {
	case s.Shell == "" && f.Kind == workflow.KindAction:
		c.add(f, s, Unsupported, SeverityWarning, s.Line, s.Col, "composite step has no shell; not analyzed")
		return
	case !supportedShell(s.Shell):
		c.add(f, s, Unsupported, SeverityWarning, s.Line, s.Col, fmt.Sprintf("step uses shell '%s'; not analyzed", s.Shell))
		return
	}
	res, err := shell.Extract(s.Run.Text, c.opts)
	if err != nil {
		c.parseError(f, s, err)
		return
	}
	for _, p := range res.Problems {
		line, col := s.Run.Position(p.Line, p.Col)
		c.add(f, s, Unsupported, SeverityWarning, line, col, p.Message)
	}
	for _, cmd := range res.Commands {
		if cmd.Kind == shell.KindTool {
			c.inlineTool(f, s, cmd)
		}
	}
}

// parseError reports a script that does not parse, at the parser's position
// when it gave one and at the run: key otherwise.
func (c *checker) parseError(f *workflow.File, s *workflow.Step, err error) {
	line, col := s.Line, s.Col
	var pe *shell.ParseError
	if errors.As(err, &pe) && pe.Line > 0 {
		line, col = s.Run.Position(pe.Line, pe.Col)
	}
	c.add(f, s, Unsupported, SeverityWarning, line, col, "cannot parse step script: "+err.Error())
}

func (c *checker) inlineTool(f *workflow.File, s *workflow.Step, cmd shell.Command) {
	line, col := s.Run.Position(cmd.Line, cmd.Col)
	msg := fmt.Sprintf("%s %s runs '%s' directly", stepRef(s), scope(f, s), cmd.Normalized)
	sev, reason := SeverityError, ""
	if cmd.Allowed {
		sev, reason = SeverityAllowed, cmd.Reason
	} else if r, ok := c.allowRule(f.Path, cmd.Normalized); ok {
		sev, reason = SeverityAllowed, r
	}
	fd := c.add(f, s, InlineTool, sev, line, col, msg)
	fd.Command, fd.Reason = cmd.Normalized, reason
}

// allowRule returns the reason of the first config allow entry matching the
// command in the given workflow.
func (c *checker) allowRule(file, command string) (string, bool) {
	for _, a := range c.cfg.Allow {
		if !config.MatchCommand(a.Match, command) {
			continue
		}
		if len(a.Workflows) == 0 {
			return a.Reason, true
		}
		for _, w := range a.Workflows {
			if config.MatchPath(w, file) {
				return a.Reason, true
			}
		}
	}
	return "", false
}

func (c *checker) add(f *workflow.File, s *workflow.Step, rule string, sev Severity, line, col int, msg string) *Finding {
	fd := Finding{
		Rule: rule, Severity: sev, File: f.Path, Line: line, Col: col,
		Job: s.Job, Step: s.Name, StepIndex: s.Index, Message: msg,
	}
	if f.Kind == workflow.KindAction {
		fd.Action = f.Path
	}
	c.findings = append(c.findings, fd)
	return &c.findings[len(c.findings)-1]
}

// supportedShell reports whether a step's shell is one the extractor parses.
func supportedShell(sh string) bool {
	return sh == "bash" || sh == "sh" || strings.HasPrefix(sh, "bash ") || strings.HasPrefix(sh, "sh ")
}

// stepRef names a step as `step "<name>"`, or `step <index>` when unnamed.
func stepRef(s *workflow.Step) string {
	if s.Name == "" {
		return fmt.Sprintf("step %d", s.Index)
	}
	return fmt.Sprintf("step %q", s.Name)
}

func scope(f *workflow.File, s *workflow.Step) string {
	if f.Kind == workflow.KindAction {
		return fmt.Sprintf("in action %q", f.Path)
	}
	return fmt.Sprintf("in job %q", s.Job)
}
