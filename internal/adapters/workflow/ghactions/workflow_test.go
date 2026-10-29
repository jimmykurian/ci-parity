// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package ghactions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadResolvesTheWorkingDirectory(t *testing.T) {
	// Arrange
	tests := []struct {
		name, rel, yaml string
		want            []string
	}{
		{
			name: "step, then job default, then workflow default",
			rel:  ".github/workflows/ci.yml",
			yaml: `defaults:
  run:
    working-directory: app
jobs:
  web:
    defaults:
      run:
        working-directory: web
    steps:
      - run: pnpm build
      - run: pnpm test
        working-directory: web/e2e
  api:
    steps:
      - run: go test ./...
`,
			want: []string{"web", "web/e2e", "app"},
		},
		{
			name: "none set",
			rel:  ".github/workflows/ci.yml",
			yaml: "jobs:\n  test:\n    steps:\n      - run: pnpm test\n",
			want: []string{""},
		},
		{
			name: "composite action step",
			rel:  ".github/actions/setup/action.yml",
			yaml: "runs:\n  using: composite\n  steps:\n    - run: pnpm install\n      shell: bash\n      working-directory: web\n",
			want: []string{"web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			root := t.TempDir()
			path := filepath.Join(root, filepath.FromSlash(tt.rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tt.yaml), 0o644); err != nil {
				t.Fatal(err)
			}

			// Act
			f, err := Load(root, tt.rel)
			if err != nil {
				t.Fatal(err)
			}

			// Assert
			var got []string
			for _, s := range f.Steps {
				got = append(got, s.WorkingDir)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("working directories (-want +got):\n%s", diff)
			}
		})
	}
}
