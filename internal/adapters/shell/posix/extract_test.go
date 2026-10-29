// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package posix

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"

	"github.com/jimmykurian/ci-parity/internal/rules"
)

var testOpts = domain.ExtractOptions{
	Tools:       set("pnpm", "npm", "go", "docker", "eslint", "python", "make"),
	Transparent: set("sudo", "env", "time", "nice", "ionice", "timeout", "nohup"),
	Invokers: [][]string{
		{"task"},
		{"go-task"},
		{"go", "tool", "task"},
		{"go", "tool", "github.com/go-task/task/v3/cmd/task"},
		{"go", "run", "github.com/go-task/task/v3/cmd/task"},
	},
}

func set(names ...string) map[string]bool {
	m := map[string]bool{}
	for _, n := range names {
		m[n] = true
	}
	return m
}

// summary renders a Result compactly: one entry per command or problem.
func summary(r domain.Extraction) []string {
	var out []string
	for _, c := range r.Commands {
		kind := "tool"
		if c.Kind == domain.KindInvoker {
			kind = "invoker"
		}
		s := fmt.Sprintf("%d:%d %s %s", c.Line, c.Col, kind, c.Normalized)
		if c.Allowed {
			s += " [allowed: " + c.Reason + "]"
		}
		out = append(out, s)
	}
	for _, p := range r.Problems {
		out = append(out, fmt.Sprintf("%d:%d problem %s", p.Line, p.Col, p.Message))
	}
	return out
}

func TestExtract(t *testing.T) {
	// Arrange
	tests := []struct {
		name, script string
		want         []string
	}{
		{"single tool", "pnpm test", []string{"1:1 tool pnpm test"}},
		{"ignored commands", "echo hi\ngit status\ngh pr view", nil},
		{"pipeline", "pnpm list | grep acme-app | sort", []string{"1:1 tool pnpm list"}},
		{"and list", "pnpm install && pnpm build || npm run fallback", []string{
			"1:1 tool pnpm install", "1:17 tool pnpm build", "1:31 tool npm run fallback",
		}},
		{"if block", "if [ -f package.json ]; then\n  pnpm lint\nelse\n  echo skip\nfi", []string{"2:3 tool pnpm lint"}},
		{"if condition", "if pnpm exec tsc --noEmit; then echo ok; fi", []string{"1:4 tool pnpm exec tsc --noEmit"}},
		{"for and while", "for d in a b; do\n  docker build $d\ndone\nwhile false; do go vet; done", []string{
			"2:3 tool docker build", "4:17 tool go vet",
		}},
		{"case", "case $X in\n  a) make all ;;\nesac", []string{"2:6 tool make all"}},
		{"function body", "build() {\n  go build ./...\n}\nbuild", []string{"2:3 tool go build ./..."}},
		{"subshell", "(cd web && pnpm build)", []string{"1:12 tool pnpm build"}},
		{"command substitution", "echo \"$(pnpm bin)\"", []string{"1:9 tool pnpm bin"}},
		{"cd then tool", "cd services/example-api && go test ./...", []string{"1:28 tool go test ./..."}},
		{"heredoc body ignored", "cat <<EOF > notes.txt\npnpm test\nEOF\npnpm lint", []string{"4:1 tool pnpm lint"}},
		{"path basename", "./node_modules/.bin/eslint .", []string{"1:1 tool ./node_modules/.bin/eslint ."}},
		{"quoted command", "'pnpm' \"test\"", []string{"1:1 tool pnpm test"}},
		{"non-literal args dropped", "pnpm test --filter $PKG", []string{"1:1 tool pnpm test --filter"}},
		{"assignment prefix", "CI=1 NODE_ENV=test pnpm test", []string{"1:20 tool pnpm test"}},
		{"assignment only", "FOO=bar", nil},
		{"env wrapper", "env FOO=1 pnpm test", []string{"1:11 tool pnpm test"}},
		{"env flags", "env -i PATH=/bin go test", []string{"1:18 tool go test"}},
		{"timeout number", "timeout 30 pnpm test", []string{"1:12 tool pnpm test"}},
		{"timeout flag and duration", "timeout --preserve-status 30s pnpm test", []string{"1:31 tool pnpm test"}},
		{"timeout flag value", "timeout -k 5 30s pnpm test", []string{"1:18 tool pnpm test"}},
		{"timeout signal", "timeout --signal KILL 30 pnpm test", []string{"1:26 tool pnpm test"}},
		{"sudo user", "sudo -u deploy docker push acme", []string{"1:16 tool docker push acme"}},
		{"sudo attached and long values", "sudo -udeploy --user=deploy docker ps", []string{"1:29 tool docker ps"}},
		{"env unset and chdir", "env -u HOME -C web pnpm build", []string{"1:20 tool pnpm build"}},
		{"env split string hides command", "env -S 'pnpm test' --flag", nil},
		{"ionice class", "ionice -c 2 -n 7 make all", []string{"1:18 tool make all"}},
		// The bash time keyword takes no -f; the flag form is the /usr/bin/time binary.
		{"time binary format", "/usr/bin/time -f %e pnpm test", []string{"1:21 tool pnpm test"}},
		{"value flag at end", "sudo -u", nil},
		{"custom wrapper gets generic rule", "chronic -v pnpm test", nil},
		{"nice with flag", "nice -n 10 make all", []string{"1:12 tool make all"}},
		{"stacked wrappers", "sudo -E env CI=1 timeout 60 docker build .", []string{"1:29 tool docker build ."}},
		{"time keyword", "time pnpm test", []string{"1:6 tool pnpm test"}},
		{"wrapper alone", "sudo -v", nil},
		{"task invoker", "task build", []string{"1:1 invoker task build"}},
		{"go tool task is an invoker, not go", "go tool task lint", []string{"1:1 invoker go tool task lint"}},
		{"go tool other is go", "go tool cover -html=c.out", []string{"1:1 tool go tool cover -html=c.out"}},
		{"go tool with task's module", "go tool github.com/go-task/task/v3/cmd/task lint", []string{
			"1:1 invoker go tool github.com/go-task/task/v3/cmd/task lint",
		}},
		{"go run with a version", "go run github.com/go-task/task/v3/cmd/task@v3.40.0 lint", []string{
			"1:1 invoker go run github.com/go-task/task/v3/cmd/task@v3.40.0 lint",
		}},
		{"go run of another module is go", "go run github.com/acme/tool@v1 lint", []string{
			"1:1 tool go run github.com/acme/tool@v1 lint",
		}},
		{"a longer path is not the module", "go run github.com/go-task/task/v3/cmd/taskfile lint", []string{
			"1:1 tool go run github.com/go-task/task/v3/cmd/taskfile lint",
		}},
		{"wrapped invoker", "env CI=1 task test", []string{"1:10 invoker task test"}},
		{"dynamic variable", "$TOOL test", []string{"1:1 problem dynamic command name '$TOOL' cannot be analyzed"}},
		{"dynamic expression", "${{ matrix.tool }} test", []string{
			"1:1 problem dynamic command name '${{ matrix.tool }}' cannot be analyzed",
		}},
		{"expression argument", "pnpm --filter ${{ inputs.pkg }} build", []string{"1:1 tool pnpm --filter build"}},
		{"column after expression", "echo ${{ github.sha }} && pnpm test", []string{"1:27 tool pnpm test"}},
		{"line after multi-line expression", "echo ${{\n  github.sha }}\npnpm test", []string{"3:1 tool pnpm test"}},
		{"expression in condition", "if [ \"${{ github.event_name }}\" = push ]; then\n  docker push acme\nfi", []string{
			"2:3 tool docker push acme",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			res, err := Extract(tt.script, testOpts)
			if err != nil {
				t.Fatal(err)
			}

			// Assert
			if diff := cmp.Diff(tt.want, summary(res)); diff != "" {
				t.Errorf("Extract(%q) (-want +got):\n%s", tt.script, diff)
			}
		})
	}
}

func TestExtractAllowComments(t *testing.T) {
	// Arrange
	tests := []struct {
		name, script string
		want         []string
	}{
		{"same line", "docker push acme # ci-parity: allow images publish here", []string{
			"1:1 tool docker push acme [allowed: images publish here]",
		}},
		{"line above", "# ci-parity: allow one-off migration\npnpm db:migrate\npnpm test", []string{
			"2:1 tool pnpm db:migrate [allowed: one-off migration]", "3:1 tool pnpm test",
		}},
		{"trailing comment does not reach next line", "pnpm a # ci-parity: allow first only\npnpm b", []string{
			"1:1 tool pnpm a [allowed: first only]", "2:1 tool pnpm b",
		}},
		{"no reason", "pnpm test # ci-parity: allow", []string{
			"1:1 tool pnpm test", "1:11 problem allow comment has no reason",
		}},
		{"not a directive", "pnpm test # ci-parity: allowed", []string{"1:1 tool pnpm test"}},
		{"expression in reason", "# ci-parity: allow runs for ${{ matrix.os }}\ngo test", []string{
			"2:1 tool go test [allowed: runs for ${{ matrix.os }}]",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			res, err := Extract(tt.script, testOpts)
			if err != nil {
				t.Fatal(err)
			}

			// Assert
			if diff := cmp.Diff(tt.want, summary(res)); diff != "" {
				t.Errorf("Extract(%q) (-want +got):\n%s", tt.script, diff)
			}
		})
	}
}

func TestExtractInvokerArgs(t *testing.T) {
	// Arrange
	want := []domain.Word{
		{Text: "--dir", Literal: true, Line: 1, Col: 6},
		{Text: "web", Literal: true, Line: 1, Col: 12},
		{Text: "build", Literal: true, Line: 1, Col: 16},
		{Text: "VERSION=${{ github.sha }}", Line: 1, Col: 22},
	}

	// Act
	res, err := Extract("task --dir web build VERSION=${{ github.sha }}", testOpts)
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if diff := cmp.Diff(want, res.Commands[0].Args); diff != "" {
		t.Errorf("Args (-want +got):\n%s", diff)
	}
}

func TestExtractParseError(t *testing.T) {
	// Arrange
	tests := []struct {
		name, script string
		want         domain.ParseError
	}{
		{"unclosed if", "echo ok\nif pnpm test; then\n  echo ${{ x }}\n", domain.ParseError{Line: 2, Col: 1, Message: "`if` statement must end with `fi`"}},
		{"position after expression", "echo ${{ github.sha }} )", domain.ParseError{Line: 1, Col: 24, Message: "a command can only contain words and redirects; encountered `)`"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := Extract(tt.script, testOpts)

			// Assert
			var pe *domain.ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("err = %v, want *domain.ParseError", err)
			}
			if diff := cmp.Diff(tt.want, *pe); diff != "" {
				t.Errorf("domain.ParseError (-want +got):\n%s", diff)
			}
		})
	}
}

var _ rules.CommandExtractor = Extractor{}
