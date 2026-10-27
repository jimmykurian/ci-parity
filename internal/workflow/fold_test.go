package workflow

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// runScript parses a one-step workflow snippet and returns the step's script.
func runScript(t *testing.T, src string) *Script {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
		t.Fatal(err)
	}
	return script(get(doc.Content[0], "run"), strings.Split(src, "\n"))
}

func TestScriptPosition(t *testing.T) {
	tests := []struct {
		name, src           string
		line, col           int // position inside the script
		wantLine, wantCol   int // position in the file
		wantFoldedPositions bool
	}{
		{"plain", "run: pnpm test\n", 1, 1, 1, 6, false},
		{"double quoted", "run: \"pnpm test\"\n", 1, 1, 1, 7, false},
		{"literal", "run: |\n  echo a\n  pnpm test\n", 2, 1, 3, 3, false},
		{"folded joins lines", "run: >\n  echo a\n  && pnpm test\n", 1, 11, 3, 6, true},
		{"folded blank line", "run: >\n  echo a\n\n  pnpm test\n", 2, 1, 4, 3, true},
		{"folded more-indented", "run: >\n  echo a\n    echo b\n  pnpm test\n", 3, 1, 4, 3, true},
		{"plain multi-line", "run: echo a\n  && pnpm test\n", 1, 11, 2, 6, true},
		{"plain with comment", "run: echo a\n  && pnpm test # note\n", 1, 11, 2, 6, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runScript(t, tt.src)
			if got := s.segs != nil; got != tt.wantFoldedPositions {
				t.Errorf("folded mapping used = %v, want %v (text %q)", got, tt.wantFoldedPositions, s.Text)
			}
			line, col := s.Position(tt.line, tt.col)
			if line != tt.wantLine || col != tt.wantCol {
				t.Errorf("Position(%d, %d) = %d:%d, want %d:%d", tt.line, tt.col, line, col, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestFoldedBlockMismatchFallsBack(t *testing.T) {
	src := []string{"run: >", "  echo a", "  echo b"}
	if segs := foldedBlock(src, 1, 2, "something else"); segs != nil {
		t.Errorf("foldedBlock with a mismatched value = %v, want nil", segs)
	}
}
