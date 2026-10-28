package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

var (
	exprPattern = regexp.MustCompile(`(?s)\$\{\{.*?\}\}`)
	wordPattern = regexp.MustCompile(`\S+`)
)

// actionInvoker checks the task names a configured action invoker step passes
// in its input.
func (c *checker) actionInvoker(f *domain.File, s *domain.Step) {
	for _, inv := range c.cfg.Invokers.Actions {
		if s.Uses != inv.Uses && !strings.HasPrefix(s.Uses, inv.Uses+"@") {
			continue
		}
		in, ok := s.With[inv.Input]
		if !ok {
			return
		}
		// Blank out expressions so their inner words are not read as task names.
		value := exprPattern.ReplaceAllStringFunc(in.Value, func(m string) string {
			return strings.Repeat("\x00", len(m))
		})
		for _, loc := range wordPattern.FindAllStringIndex(value, -1) {
			name := value[loc[0]:loc[1]]
			if strings.Contains(name, "\x00") {
				continue
			}
			line, col := in.Line, in.Col
			if !strings.Contains(in.Value, "\n") {
				col += loc[0]
			}
			c.task(f, s, name, line, col)
		}
		return
	}
}

// task reports P002 for a name the runner does not define and P003 for an
// internal one.
func (c *checker) task(f *domain.File, s *domain.Step, name string, line, col int) {
	switch {
	case !c.tasks[name]:
		msg := fmt.Sprintf("'%s %s' is not defined in %s", c.runner.Command(), name, c.runner.File())
		fd := c.add(f, s, domain.UnknownTask, domain.SeverityError, line, col, msg)
		if closest := c.suggest.closest(name); closest != "" {
			fd.Message += fmt.Sprintf(" (closest: %s)", closest)
			fd.Suggestion = c.runner.Command() + " " + closest
		}
	case c.runner.IsInternal(name):
		c.add(f, s, domain.InternalTask, domain.SeverityError, line, col,
			fmt.Sprintf("'%s' is marked internal: true and should not be invoked from CI", name))
	}
}
