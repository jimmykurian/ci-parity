package shell

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var testOpts = Options{
	Tools:       set("pnpm", "npm", "go", "docker", "eslint", "python", "make"),
	Transparent: set("sudo", "env", "time", "nice", "timeout", "nohup"),
	Invokers:    [][]string{{"task"}, {"go-task"}, {"go", "tool", "task"}},
}

func set(names ...string) map[string]bool {
	m := map[string]bool{}
	for _, n := range names {
		m[n] = true
	}
	return m
}

// summary renders a Result compactly: one entry per command or problem.
func summary(r Result) []string {
	var out []string
	for _, c := range r.Commands {
		kind := "tool"
		if c.Kind == KindInvoker {
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
	tests := []struct {
		name, script string
		want         []string
	}{
		{"single tool", "pnpm test", []string{"1:1 tool pnpm test"}},
		{"ignored commands", "echo hi\ngit status\ngh pr view", nil},
		{"pipeline", "pnpm list | grep acme-app | sort", []string{"1:1 tool pnpm list"}},
		{"and list", "pnpm install && pnpm build || npm run fallback", []string{
			"1:1 tool pnpm install", "1:17 tool pnpm build", "1:31 tool npm run fallback"}},
		{"if block", "if [ -f package.json ]; then\n  pnpm lint\nelse\n  echo skip\nfi", []string{"2:3 tool pnpm lint"}},
		{"if condition", "if pnpm exec tsc --noEmit; then echo ok; fi", []string{"1:4 tool pnpm exec tsc --noEmit"}},
		{"for and while", "for d in a b; do\n  docker build $d\ndone\nwhile false; do go vet; done", []string{
			"2:3 tool docker build", "4:17 tool go vet"}},
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
		// Known limitation per SPEC 5.3: wrapper flags that take a separate value are not
		// understood, so the value is read as the command.
		{"wrapper flag value", "timeout -k 5 30s pnpm test", nil},
		{"nice with flag", "nice -n 10 make all", []string{"1:12 tool make all"}},
		{"stacked wrappers", "sudo -E env CI=1 timeout 60 docker build .", []string{"1:29 tool docker build ."}},
		{"time keyword", "time pnpm test", []string{"1:6 tool pnpm test"}},
		{"wrapper alone", "sudo -v", nil},
		{"task invoker", "task build", []string{"1:1 invoker task build"}},
		{"go tool task is an invoker, not go", "go tool task lint", []string{"1:1 invoker go tool task lint"}},
		{"go tool other is go", "go tool cover -html=c.out", []string{"1:1 tool go tool cover -html=c.out"}},
		{"wrapped invoker", "env CI=1 task test", []string{"1:10 invoker task test"}},
		{"dynamic variable", "$TOOL test", []string{"1:1 problem dynamic command name '$TOOL' cannot be analyzed"}},
		{"dynamic expression", "${{ matrix.tool }} test", []string{
			"1:1 problem dynamic command name '${{ matrix.tool }}' cannot be analyzed"}},
		{"expression argument", "pnpm --filter ${{ inputs.pkg }} build", []string{"1:1 tool pnpm --filter build"}},
		{"expression in condition", "if [ \"${{ github.event_name }}\" = push ]; then\n  docker push acme\nfi", []string{
			"2:3 tool docker push acme"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Extract(tt.script, testOpts)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, summary(res)); diff != "" {
				t.Errorf("Extract(%q) (-want +got):\n%s", tt.script, diff)
			}
		})
	}
}

func TestExtractAllowComments(t *testing.T) {
	tests := []struct {
		name, script string
		want         []string
	}{
		{"same line", "docker push acme # ci-parity: allow images publish here", []string{
			"1:1 tool docker push acme [allowed: images publish here]"}},
		{"line above", "# ci-parity: allow one-off migration\npnpm db:migrate\npnpm test", []string{
			"2:1 tool pnpm db:migrate [allowed: one-off migration]", "3:1 tool pnpm test"}},
		{"trailing comment does not reach next line", "pnpm a # ci-parity: allow first only\npnpm b", []string{
			"1:1 tool pnpm a [allowed: first only]", "2:1 tool pnpm b"}},
		{"no reason", "pnpm test # ci-parity: allow", []string{
			"1:1 tool pnpm test", "1:11 problem allow comment has no reason"}},
		{"not a directive", "pnpm test # ci-parity: allowed", []string{"1:1 tool pnpm test"}},
		{"expression in reason", "# ci-parity: allow runs for ${{ matrix.os }}\ngo test", []string{
			"2:1 tool go test [allowed: runs for ${{ matrix.os }}]"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Extract(tt.script, testOpts)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, summary(res)); diff != "" {
				t.Errorf("Extract(%q) (-want +got):\n%s", tt.script, diff)
			}
		})
	}
}

func TestExtractInvokerArgs(t *testing.T) {
	res, err := Extract("task --dir web build VERSION=${{ github.sha }}", testOpts)
	if err != nil {
		t.Fatal(err)
	}
	want := []Word{{"--dir", true}, {"web", true}, {"build", true}, {"VERSION=${{ github.sha }}", false}}
	if diff := cmp.Diff(want, res.Commands[0].Args); diff != "" {
		t.Errorf("Args (-want +got):\n%s", diff)
	}
}

func TestExtractParseError(t *testing.T) {
	_, err := Extract("if pnpm test; then\n  echo ${{ x }}\n", testOpts)
	if err == nil {
		t.Fatal("want parse error")
	}
	if strings.Contains(err.Error(), "__GHA_EXPR_") {
		t.Errorf("error leaks placeholder: %v", err)
	}
}
