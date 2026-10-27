// Package shell extracts tool and task-runner invocations from run: scripts.
package shell

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// Kind classifies an extracted command.
type Kind int

const (
	// KindTool is a call to a command on the tools list.
	KindTool Kind = iota + 1
	// KindInvoker is a call to the task runner through a shell invoker.
	KindInvoker
)

// Options configures extraction.
type Options struct {
	Tools       map[string]bool // tool basenames to report
	Transparent map[string]bool // wrapper basenames to look through
	Invokers    [][]string      // shell invokers as word sequences, e.g. {"go", "tool", "task"}
}

// Word is one argument of a command.
type Word struct {
	Text    string // literal value, or the source text when not literal
	Literal bool
}

// Command is one tool or invoker call found in a script.
type Command struct {
	Kind       Kind
	Name       string // tool basename, or the invoker as configured
	Normalized string // literal words from the command on, joined by single spaces
	Args       []Word // invoker arguments after the invoker words
	Line, Col  int    // 1-based position within the script
	Allowed    bool
	Reason     string
}

// Problem is a condition inside a script that prevents analysis (rule P005).
type Problem struct {
	Line, Col int
	Message   string
}

// Result holds everything extracted from one script, sorted by position.
type Result struct {
	Commands []Command
	Problems []Problem
}

var exprPattern = regexp.MustCompile(`(?s)\$\{\{.*?\}\}`)

// Extract parses script as bash and returns the tool and invoker calls in it.
// A parse error is returned as an error; the caller reports it as P005.
func Extract(script string, opts Options) (Result, error) {
	e := &extractor{opts: opts, allows: map[int]allowComment{}}
	src := e.replaceExpressions(script)
	e.lines = strings.Split(src, "\n")
	parser := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(syntax.LangBash))
	file, err := parser.Parse(strings.NewReader(src), "")
	if err != nil {
		return Result{}, fmt.Errorf("%s", e.restore(err.Error()))
	}
	syntax.Walk(file, e.visit)
	e.applyAllows()
	sort.SliceStable(e.res.Commands, func(i, j int) bool {
		return before(e.res.Commands[i].Line, e.res.Commands[i].Col, e.res.Commands[j].Line, e.res.Commands[j].Col)
	})
	sort.SliceStable(e.res.Problems, func(i, j int) bool {
		return before(e.res.Problems[i].Line, e.res.Problems[i].Col, e.res.Problems[j].Line, e.res.Problems[j].Col)
	})
	return e.res, nil
}

func before(l1, c1, l2, c2 int) bool {
	if l1 != l2 {
		return l1 < l2
	}
	return c1 < c2
}

type extractor struct {
	opts   Options
	exprs  []string
	lines  []string
	allows map[int]allowComment
	res    Result
}

// replaceExpressions swaps each ${{ }} for a token the shell parser treats as
// an ordinary word, remembering the original text for messages.
func (e *extractor) replaceExpressions(script string) string {
	return exprPattern.ReplaceAllStringFunc(script, func(m string) string {
		e.exprs = append(e.exprs, m)
		return fmt.Sprintf("__GHA_EXPR_%d__", len(e.exprs)-1)
	})
}

// restore puts the original ${{ }} text back into s.
func (e *extractor) restore(s string) string {
	for i := len(e.exprs) - 1; i >= 0; i-- {
		s = strings.ReplaceAll(s, fmt.Sprintf("__GHA_EXPR_%d__", i), e.exprs[i])
	}
	return s
}

func (e *extractor) visit(n syntax.Node) bool {
	switch n := n.(type) {
	case *syntax.Redirect:
		// Heredoc bodies are data, not commands; only the redirect target is walked.
		if n.Word != nil {
			syntax.Walk(n.Word, e.visit)
		}
		return false
	case *syntax.Comment:
		e.comment(n)
	case *syntax.CallExpr:
		e.call(n)
	}
	return true
}

func (e *extractor) call(c *syntax.CallExpr) {
	words := make([]Word, len(c.Args))
	for i, a := range c.Args {
		words[i] = e.word(a)
	}
	i := e.peel(words)
	if i >= len(words) {
		return
	}
	pos := c.Args[i].Pos()
	line, col := int(pos.Line()), int(pos.Col())
	cmd := words[i]
	if !cmd.Literal {
		e.problem(line, col, fmt.Sprintf("dynamic command name '%s' cannot be analyzed", cmd.Text))
		return
	}
	if inv, n := e.matchInvoker(words[i:]); n > 0 {
		e.res.Commands = append(e.res.Commands, Command{
			Kind: KindInvoker, Name: inv, Normalized: normalize(words[i:]),
			Args: words[i+n:], Line: line, Col: col,
		})
		return
	}
	if name := path.Base(cmd.Text); e.opts.Tools[name] {
		e.res.Commands = append(e.res.Commands, Command{
			Kind: KindTool, Name: name, Normalized: normalize(words[i:]), Line: line, Col: col,
		})
	}
}

// word returns the literal value of w, or its source text when w contains an
// expansion, a substitution or a ${{ }} expression.
func (e *extractor) word(w *syntax.Word) Word {
	var b strings.Builder
	for _, part := range w.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			b.WriteString(p.Value)
		case *syntax.SglQuoted:
			b.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, q := range p.Parts {
				lit, ok := q.(*syntax.Lit)
				if !ok {
					return Word{Text: e.source(w)}
				}
				b.WriteString(lit.Value)
			}
		default:
			return Word{Text: e.source(w)}
		}
	}
	if strings.Contains(b.String(), "__GHA_EXPR_") {
		return Word{Text: e.source(w)}
	}
	return Word{Text: b.String(), Literal: true}
}

func (e *extractor) source(w *syntax.Word) string {
	var b strings.Builder
	if err := syntax.NewPrinter().Print(&b, w); err != nil {
		return "?"
	}
	return e.restore(b.String())
}

// peel skips transparent wrappers such as sudo, env and timeout, along with
// their flags, assignments and (for timeout and nice) a leading number. It
// returns the index of the real command word.
func (e *extractor) peel(words []Word) int {
	i := 0
	for i < len(words) && words[i].Literal && e.opts.Transparent[path.Base(words[i].Text)] {
		numeric := path.Base(words[i].Text) == "timeout" || path.Base(words[i].Text) == "nice"
		i++
		for i < len(words) && words[i].Literal {
			t := words[i].Text
			if strings.HasPrefix(t, "-") || strings.Contains(t, "=") {
				i++
				continue
			}
			if numeric && t != "" && t[0] >= '0' && t[0] <= '9' {
				numeric = false
				i++
				continue
			}
			break
		}
	}
	return i
}

// matchInvoker returns the longest configured invoker that words begin with,
// and how many words it spans.
func (e *extractor) matchInvoker(words []Word) (string, int) {
	best, bestLen := "", 0
	for _, inv := range e.opts.Invokers {
		if len(inv) == 0 || len(inv) > len(words) || len(inv) <= bestLen {
			continue
		}
		if path.Base(words[0].Text) != inv[0] {
			continue
		}
		ok := true
		for k := 1; k < len(inv); k++ {
			if !words[k].Literal || words[k].Text != inv[k] {
				ok = false
				break
			}
		}
		if ok {
			best, bestLen = strings.Join(inv, " "), len(inv)
		}
	}
	return best, bestLen
}

func (e *extractor) problem(line, col int, msg string) {
	e.res.Problems = append(e.res.Problems, Problem{Line: line, Col: col, Message: msg})
}

func normalize(words []Word) string {
	lits := make([]string, 0, len(words))
	for _, w := range words {
		if w.Literal {
			lits = append(lits, w.Text)
		}
	}
	return strings.Join(lits, " ")
}
