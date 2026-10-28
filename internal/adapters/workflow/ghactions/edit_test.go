package ghactions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

func TestApply(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "ci.yml")
	in := "steps:\r\n  - run: pnpm test # note\r\n  - run: pnpm lint\r\n  - run: énv pnpm x\r\n"
	if err := os.WriteFile(path, []byte(in), 0o600); err != nil {
		t.Fatal(err)
	}
	edits := []domain.Edit{
		{File: "ci.yml", Line: 2, Col: 10, Old: "pnpm test", New: "task test"},
		{File: "ci.yml", Line: 3, Col: 10, Old: "pnpm build", New: "task build"}, // text moved: skipped
		{File: "ci.yml", Line: 4, Col: 14, Old: "pnpm x", New: "task x"},         // after a multibyte character; yaml.v3 columns count characters
		{File: "ci.yml", Line: 9, Col: 1, Old: "x", New: "y"},                    // no such line
	}
	applied, err := Source{Root: root}.Apply("ci.yml", edits)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]domain.Edit{edits[0], edits[2]}, applied); diff != "" {
		t.Errorf("applied (-want +got):\n%s", diff)
	}
	want := "steps:\r\n  - run: task test # note\r\n  - run: pnpm lint\r\n  - run: énv task x\r\n"
	got, _ := os.ReadFile(path)
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("file (-want +got):\n%s", diff)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestApplyRefusesPathsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	if _, err := (Source{Root: filepath.Join(root, "repo")}).Apply("../outside.yml", []domain.Edit{{Line: 1, Col: 1, Old: "a", New: "b"}}); err == nil {
		t.Error("an edit outside the repository was accepted")
	}
}
