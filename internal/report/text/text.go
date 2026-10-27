// Package text renders findings as human-readable text grouped by file.
package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/jimmykurian/ci-parity/internal/rules"
)

// nameWidth fits the longest rule name, "internal-task".
const nameWidth = 13

// Write renders sorted findings to w. Tolerated and allowed findings are shown
// only when verbose is set; the summary line always counts them.
func Write(w io.Writer, findings []rules.Finding, verbose bool) error {
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
	b.WriteString(summary(findings) + "\n")
	_, err := io.WriteString(w, b.String())
	return err
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
