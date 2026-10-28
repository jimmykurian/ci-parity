package makefile

import (
	"strconv"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Command is how CI calls a make target.
func (m *Makefile) Command() string { return "make" }

// otherFile are make flags that point the call at another makefile or
// directory, whose targets cannot be checked.
var otherFile = map[string]bool{"-f": true, "--file": true, "--makefile": true, "-C": true, "--directory": true}

// valueFlags take the next argument as their value.
var valueFlags = map[string]bool{
	"-I": true, "--include-dir": true, "-o": true, "--old-file": true, "--assume-old": true,
	"-W": true, "--what-if": true, "--new-file": true, "--assume-new": true,
}

// optionalNumber take the next argument only when it is a number: `make -j 4`
// but `make -j build`.
var optionalNumber = map[string]bool{"-j": true, "--jobs": true, "-l": true, "--load-average": true, "--max-load": true}

// Names returns the targets a make call names. It returns none when the call
// selects another makefile or directory, or names no target and so runs the
// default goal. Variables (NAME=value) and words that are not literal are
// skipped.
func (m *Makefile) Names(args []domain.Word) []domain.Word {
	var names []domain.Word
	flags := true
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case !a.Literal:
		case flags && a.Text == "--":
			flags = false
		case flags && selectsOtherFile(a.Text):
			return nil
		case flags && valueFlags[a.Text]:
			i++
		case flags && optionalNumber[a.Text]:
			if i+1 < len(args) && isNumber(args[i+1].Text) {
				i++
			}
		case flags && strings.HasPrefix(a.Text, "-"):
		case strings.Contains(a.Text, "="):
		default:
			names = append(names, a)
		}
	}
	return names
}

// selectsOtherFile reports whether a flag, in any of its forms (-f x, -fx,
// --file=x), points make at another makefile or directory.
func selectsOtherFile(arg string) bool {
	if otherFile[arg] {
		return true
	}
	for _, long := range []string{"--file=", "--makefile=", "--directory="} {
		if strings.HasPrefix(arg, long) {
			return true
		}
	}
	return len(arg) > 2 && !strings.HasPrefix(arg, "--") && (strings.HasPrefix(arg, "-f") || strings.HasPrefix(arg, "-C"))
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
