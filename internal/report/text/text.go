// Package text renders findings as human-readable text grouped by file.
package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/baseline"
	"github.com/jimmykurian/ci-parity/internal/rules"
)

// nameWidth fits the longest rule name, "internal-task".
const nameWidth = 13

// Write renders sorted findings and the ratchet rows to w. Tolerated and
// allowed findings are shown only when verbose is set; the summary line always
// counts them.
func Write(w io.Writer, findings []rules.Finding, rows []baseline.Row, verbose bool) error {
	var b strings.Builder
	var shown []rules.Finding
	for _, f := range findings {
		if verbose || f.Severity == rules.SeverityError || f.Severity == rules.SeverityWarning {
			shown = append(shown, f)
		}
	}
	posWidth := 0
	for _, f := range shown {
		posWidth = max(posWidth, len(position(f)))
	}
	for i, f := range shown {
		if i == 0 || f.File != shown[i-1].File {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(f.File + "\n")
		}
		fmt.Fprintf(&b, "  %-*s%s %-*s %s%s\n", posWidth+3, position(f), f.Rule, nameWidth, f.Name(), f.Message, label(f))
	}
	if len(shown) > 0 {
		b.WriteString("\n")
	}
	if r := ratchet(rows); r != "" {
		b.WriteString(r + "\n")
	}
	b.WriteString(summary(findings) + "\n")
	_, err := io.WriteString(w, b.String())
	return err
}

// ratchet renders files whose count moved off their baseline: improvements to
// lock in and regressions. It returns "" when there are none.
func ratchet(rows []baseline.Row) string {
	var lines [][4]string // file, baseline state, current count, note
	for _, r := range rows {
		if !r.Improved() && !r.Regressed() {
			continue
		}
		state, note := fmt.Sprintf("baseline %d", r.Baseline), "run `ci-parity baseline` to lock in"
		if r.Locked() {
			state = "locked"
		}
		if r.Regressed() {
			note = "REGRESSION"
		}
		lines = append(lines, [4]string{r.File, state, fmt.Sprintf("now %d", r.Current), note})
	}
	if len(lines) == 0 {
		return ""
	}
	var widths [3]int
	for _, l := range lines {
		for i := range widths {
			widths[i] = max(widths[i], len(l[i]))
		}
	}
	var b strings.Builder
	b.WriteString("ratchet\n")
	for _, l := range lines {
		fmt.Fprintf(&b, "  %-*s  %-*s   %-*s   %s\n", widths[0], l[0], widths[1], l[1], widths[2], l[2], l[3])
	}
	return b.String()
}

func position(f rules.Finding) string {
	return fmt.Sprintf("%d:%d", f.Line, f.Col)
}

func label(f rules.Finding) string {
	switch f.Severity {
	case rules.SeverityTolerated:
		return " [tolerated]"
	case rules.SeverityAllowed:
		return " [allowed: " + f.Reason + "]"
	}
	return ""
}

func summary(findings []rules.Finding) string {
	counts := map[rules.Severity]int{}
	for _, f := range findings {
		counts[f.Severity]++
	}
	return fmt.Sprintf("%s, %s, %d tolerated, %d allowed",
		plural(counts[rules.SeverityError], "error"),
		plural(counts[rules.SeverityWarning], "warning"),
		counts[rules.SeverityTolerated], counts[rules.SeverityAllowed])
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
