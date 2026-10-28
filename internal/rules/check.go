// Package rules evaluates the ci-parity rules over loaded workflows and
// produces findings shared by every reporter.
package rules

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Check evaluates every step of files against cfg and the task runner r,
// using ex to read run: scripts, and returns sorted findings.
func Check(cfg *domain.Config, r TaskSource, ex CommandExtractor, files []*domain.File) []domain.Finding {
	c := checker{cfg: cfg, opts: shellOptions(cfg), runner: r, extract: ex, tasks: map[string]bool{}, suggest: newSuggester(r)}
	for _, name := range r.Tasks() {
		c.tasks[name] = true
	}
	for _, f := range files {
		for i := range f.Steps {
			c.step(f, &f.Steps[i])
		}
	}
	domain.Sort(c.findings)
	return c.findings
}

func shellOptions(cfg *domain.Config) domain.ExtractOptions {
	opts := domain.ExtractOptions{Tools: cfg.ToolSet(), Transparent: map[string]bool{}}
	for _, t := range cfg.Transparent {
		opts.Transparent[t] = true
	}
	for _, inv := range cfg.Invokers.Shell {
		opts.Invokers = append(opts.Invokers, strings.Fields(inv))
	}
	return opts
}

type checker struct {
	cfg      *domain.Config
	opts     domain.ExtractOptions
	runner   TaskSource
	extract  CommandExtractor
	tasks    map[string]bool
	suggest  *suggester
	findings []domain.Finding
}

func (c *checker) step(f *domain.File, s *domain.Step) {
	if s.Run == nil {
		if s.Uses != "" {
			c.actionInvoker(f, s)
		}
		return
	}
	switch {
	case s.Shell == "" && f.Kind == domain.KindAction:
		c.add(f, s, domain.Unsupported, domain.SeverityWarning, s.Line, s.Col, "composite step has no shell; not analyzed")
		return
	case !c.extract.Supports(s.Shell):
		c.add(f, s, domain.Unsupported, domain.SeverityWarning, s.Line, s.Col, fmt.Sprintf("step uses shell '%s'; not analyzed", s.Shell))
		return
	}
	res, err := c.extract.Extract(s.Run.Text, c.opts)
	if err != nil {
		c.parseError(f, s, err)
		return
	}
	for _, p := range res.Problems {
		line, col := s.Run.Position(p.Line, p.Col)
		c.add(f, s, domain.Unsupported, domain.SeverityWarning, line, col, p.Message)
	}
	for _, cmd := range res.Commands {
		if cmd.Kind == domain.KindTool {
			c.inlineTool(f, s, cmd)
			continue
		}
		for _, w := range c.runner.Names(cmd.Args) {
			line, col := s.Run.Position(w.Line, w.Col)
			c.task(f, s, w.Text, line, col)
		}
	}
}

// parseError reports a script that does not parse, at the parser's position
// when it gave one and at the run: key otherwise.
func (c *checker) parseError(f *domain.File, s *domain.Step, err error) {
	line, col := s.Line, s.Col
	var pe *domain.ParseError
	if errors.As(err, &pe) && pe.Line > 0 {
		line, col = s.Run.Position(pe.Line, pe.Col)
	}
	c.add(f, s, domain.Unsupported, domain.SeverityWarning, line, col, "cannot parse step script: "+err.Error())
}

func (c *checker) inlineTool(f *domain.File, s *domain.Step, cmd domain.Command) {
	line, col := s.Run.Position(cmd.Line, cmd.Col)
	msg := fmt.Sprintf("%s %s runs '%s' directly", stepRef(s), scope(f, s), cmd.Normalized)
	sev, reason := domain.SeverityError, ""
	if cmd.Allowed {
		sev, reason = domain.SeverityAllowed, cmd.Reason
	} else if r, ok := c.allowRule(f.Path, cmd.Normalized); ok {
		sev, reason = domain.SeverityAllowed, r
	}
	suggestion := c.suggest.command(cmd.Normalized)
	if suggestion != "" {
		msg += fmt.Sprintf("; already available as '%s'", suggestion)
	}
	fd := c.add(f, s, domain.InlineTool, sev, line, col, msg)
	fd.Command, fd.Reason, fd.Suggestion = cmd.Normalized, reason, suggestion
	if suggestion != "" && sev != domain.SeverityAllowed {
		fd.Fix = fixEdit(f.Path, s.Run, cmd, suggestion)
	}
}

// unsafePrefix are characters that, before the command, mean more than
// wrappers and variable assignments: operators, redirects, quoting,
// expansions and comments.
const unsafePrefix = ";&|<>()`$'\"\\{}#"

// fixEdit returns the edit that makes a step call the suggested task, or nil
// when that could change what the step does. The step's value must sit on one
// line; whatever precedes the command may only be transparent wrappers with
// their options and NAME=value assignments, which stay, so `sudo pnpm lint`
// becomes `sudo task lint`; and from the command to the end of the value must
// be exactly the command, with no operator, quoting or expansion.
func fixEdit(file string, run *domain.Script, cmd domain.Command, suggestion string) *domain.Edit {
	value := strings.TrimSuffix(run.Value, "\n")
	if !run.SingleLine || cmd.Line != 1 || cmd.Col < 1 || cmd.Col-1 > len(value) || strings.Contains(value, "${{") {
		return nil
	}
	prefix, rest := value[:cmd.Col-1], value[cmd.Col-1:] // shell columns count bytes
	if strings.ContainsAny(prefix, unsafePrefix) || strings.Join(strings.Fields(rest), " ") != cmd.Normalized {
		return nil
	}
	col := run.Col
	if run.Block {
		col = run.Indent + 1
	}
	// File columns count characters.
	return &domain.Edit{File: file, Line: run.Line, Col: col + utf8.RuneCountInString(prefix), Old: rest, New: suggestion}
}

// allowRule returns the reason of the first config allow entry matching the
// command in the given workflow.
func (c *checker) allowRule(file, command string) (string, bool) {
	for _, a := range c.cfg.Allow {
		if !domain.MatchCommand(a.Match, command) {
			continue
		}
		if len(a.Workflows) == 0 {
			return a.Reason, true
		}
		for _, w := range a.Workflows {
			if domain.MatchPath(w, file) {
				return a.Reason, true
			}
		}
	}
	return "", false
}

func (c *checker) add(f *domain.File, s *domain.Step, rule string, sev domain.Severity, line, col int, msg string) *domain.Finding {
	fd := domain.Finding{
		Rule: rule, Severity: sev, File: f.Path, Line: line, Col: col,
		Job: s.Job, Step: s.Name, StepIndex: s.Index, Message: msg,
	}
	if f.Kind == domain.KindAction {
		fd.Action = f.Path
	}
	c.findings = append(c.findings, fd)
	return &c.findings[len(c.findings)-1]
}

// stepRef names a step as `step "<name>"`, or `step <index>` when unnamed.
func stepRef(s *domain.Step) string {
	if s.Name == "" {
		return fmt.Sprintf("step %d", s.Index)
	}
	return fmt.Sprintf("step %q", s.Name)
}

func scope(f *domain.File, s *domain.Step) string {
	if f.Kind == domain.KindAction {
		return fmt.Sprintf("in action %q", f.Path)
	}
	return fmt.Sprintf("in job %q", s.Job)
}
