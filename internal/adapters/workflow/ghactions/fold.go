// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package ghactions

import (
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// foldLine is one source line of a folded scalar. An empty text is a blank line.
type foldLine struct {
	text      string
	line, col int
}

// fold joins lines the way YAML folds them: a break between two text lines
// becomes a space, blank lines become newlines, and breaks around more-indented
// lines are kept.
func fold(lines []foldLine) (string, []domain.Segment) {
	var b strings.Builder
	var segs []domain.Segment
	prevSpaced, started := false, false
	empty := 0
	for _, l := range lines {
		if l.text == "" {
			empty++
			continue
		}
		spaced := l.text[0] == ' ' || l.text[0] == '\t'
		switch {
		case !started:
			b.WriteString(strings.Repeat("\n", empty))
		case !prevSpaced && !spaced && empty == 0:
			b.WriteByte(' ')
		case !prevSpaced && !spaced:
			b.WriteString(strings.Repeat("\n", empty))
		default:
			b.WriteString(strings.Repeat("\n", empty+1))
		}
		segs = append(segs, domain.Segment{Off: b.Len(), Line: l.line, Col: l.col})
		b.WriteString(l.text)
		prevSpaced, started, empty = spaced, true, 0
	}
	return b.String(), segs
}

// foldedBlock maps a `>` block scalar whose header is on line header (1-based)
// and whose content is indented by indent. It returns nil when the rebuilt
// folding does not reproduce value, so callers fall back to line counting.
func foldedBlock(src []string, header, indent int, value string) []domain.Segment {
	var lines []foldLine
	for i := header; i < len(src); i++ {
		l := src[i]
		if strings.TrimSpace(l) == "" {
			lines = append(lines, foldLine{line: i + 1})
			continue
		}
		if len(l)-len(strings.TrimLeft(l, " ")) < indent {
			break
		}
		lines = append(lines, foldLine{text: strings.TrimRight(l[indent:], " \t\r"), line: i + 1, col: indent + 1})
	}
	return verified(lines, value)
}

// foldedPlain maps a plain scalar that starts at line and col (1-based) and
// continues over several lines. It returns nil for a single-line value or when
// it cannot reproduce value.
func foldedPlain(src []string, line, col int, value string) []domain.Segment {
	if !strings.Contains(value, " ") || line > len(src) {
		return nil
	}
	var lines []foldLine
	for i := line - 1; i < len(src); i++ {
		text, start := src[i], 0
		if i == line-1 {
			text, start = text[min(col-1, len(text)):], col-1
		}
		if c := strings.Index(text, " #"); c >= 0 {
			text = text[:c]
		}
		trimmed := strings.TrimSpace(text)
		lead := strings.Index(text, trimmed)
		lines = append(lines, foldLine{text: trimmed, line: i + 1, col: start + lead + 1})
		if segs := verified(lines, value); segs != nil {
			if len(lines) == 1 {
				return nil // single-line values map directly
			}
			return segs
		}
		if len(lines) > strings.Count(value, " ")+strings.Count(value, "\n")+1 {
			return nil
		}
	}
	return nil
}

func verified(lines []foldLine, value string) []domain.Segment {
	text, segs := fold(lines)
	if strings.TrimRight(text, "\n") != strings.TrimRight(value, "\n") {
		return nil
	}
	return segs
}
