package taskfile

import (
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Command is how CI calls a Taskfile task.
func (t *Taskfile) Command() string { return "task" }

// Internal is how P003 describes an internal task.
func (t *Taskfile) Internal() string { return "is marked internal: true" }

// valueFlags are task flags that take the next argument as their value (docs/reference/analysis.md, Invokers).
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

// Names returns the task names among the arguments of a shell invoker call. It
// returns none when the call selects another Taskfile. Names that are not
// literal, such as ${{ inputs.task }}, cannot be checked and are skipped.
func (t *Taskfile) Names(args []domain.Word) []domain.Word {
	var names []domain.Word
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
