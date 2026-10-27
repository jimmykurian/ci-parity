package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/shell"
	"github.com/jimmykurian/ci-parity/internal/workflow"
)

// valueFlags are task flags that take the next argument as their value (SPEC 5.4).
var valueFlags = map[string]bool{
	"-t": true, "--taskfile": true, "-d": true, "--dir": true,
	"-C": true, "--concurrency": true, "-I": true, "--interval": true, "-o": true, "--output": true,
}

// otherTaskfile reports whether a task flag points the call at a Taskfile
// other than the root one, whose tasks ci-parity cannot check.
func otherTaskfile(arg string) bool {
	switch arg {
	case "-t", "--taskfile", "-d", "--dir", "-g", "--global":
		return true
	}
	return strings.HasPrefix(arg, "--taskfile=") || strings.HasPrefix(arg, "--dir=") ||
		(len(arg) > 2 && (strings.HasPrefix(arg, "-t") || strings.HasPrefix(arg, "-d")) && !strings.HasPrefix(arg, "--"))
}

// taskNames returns the task names among the arguments of a shell invoker. It
// returns none when the call selects another Taskfile. Names that are not
// literal, such as ${{ inputs.task }}, cannot be checked and are skipped.
func taskNames(args []shell.Word) []shell.Word {
	var names []shell.Word
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case !a.Literal:
		case a.Text == "--":
			return names
		case otherTaskfile(a.Text):
			return nil
		case valueFlags[a.Text]:
			i++
		case strings.HasPrefix(a.Text, "-"), strings.Contains(a.Text, "="):
		default:
			names = append(names, a)
		}
	}
	return names
}

var (
	exprPattern = regexp.MustCompile(`(?s)\$\{\{.*?\}\}`)
	wordPattern = regexp.MustCompile(`\S+`)
)

// actionInvoker checks the task names a configured action invoker step passes
// in its input.
func (c *checker) actionInvoker(f *workflow.File, s *workflow.Step) {
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
func (c *checker) task(f *workflow.File, s *workflow.Step, name string, line, col int) {
	switch {
	case !c.tasks[name]:
		msg := fmt.Sprintf("'task %s' is not defined in %s", name, c.runner.File())
		closest := c.suggest.closest(name)
		if closest != "" {
			msg += fmt.Sprintf(" (closest: %s)", closest)
		}
		fd := c.add(f, s, UnknownTask, SeverityError, line, col, msg)
		fd.Suggestion = closest
	case c.runner.IsInternal(name):
		c.add(f, s, InternalTask, SeverityError, line, col,
			fmt.Sprintf("'%s' is marked internal: true and should not be invoked from CI", name))
	}
}
