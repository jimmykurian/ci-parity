package shell

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

var allowPattern = regexp.MustCompile(`^\s*ci-parity:\s*allow(?:\s+(.*?))?\s*$`)

// allowComment is a `# ci-parity: allow <reason>` comment.
type allowComment struct {
	reason string
	alone  bool // the comment is the only thing on its line
}

func (e *extractor) comment(c *syntax.Comment) {
	m := allowPattern.FindStringSubmatch(c.Text)
	if m == nil {
		return
	}
	line, col := int(c.Hash.Line()), int(c.Hash.Col())
	reason := e.restore(strings.TrimSpace(m[1]))
	if reason == "" {
		e.problem(line, col, "allow comment has no reason")
		return
	}
	alone := line <= len(e.lines) && strings.TrimSpace(e.lines[line-1][:col-1]) == ""
	e.allows[line] = allowComment{reason: reason, alone: alone}
}

// applyAllows marks commands allowed by a comment on the same line or by a
// comment alone on the line directly above.
func (e *extractor) applyAllows() {
	for i := range e.res.Commands {
		c := &e.res.Commands[i]
		if a, ok := e.allows[c.Line]; ok {
			c.Allowed, c.Reason = true, a.reason
		} else if a, ok := e.allows[c.Line-1]; ok && a.alone {
			c.Allowed, c.Reason = true, a.reason
		}
	}
}
