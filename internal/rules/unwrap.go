package rules

import "strings"

// runner is a package runner that suggestions look through (SPEC 5.6).
type runner struct {
	words      []string        // the runner's leading words
	valueFlags map[string]bool // options that take the next argument as their value
	opaque     map[string]bool // options that hide the command, so nothing is unwrapped
}

var npxFlags = map[string]bool{"-p": true, "--package": true}

var uvFlags = map[string]bool{
	"--with": true, "--with-editable": true, "--with-requirements": true, "--python": true,
	"-p": true, "--project": true, "--directory": true, "--extra": true, "--group": true,
	"--package": true, "--env-file": true, "--from": true, "--index": true, "--default-index": true,
}

var npxOpaque = map[string]bool{"-c": true, "--call": true}

var runners = []runner{
	{words: []string{"npx"}, valueFlags: npxFlags, opaque: npxOpaque},
	{words: []string{"pnpx"}, valueFlags: npxFlags, opaque: npxOpaque},
	{words: []string{"pnx"}, valueFlags: npxFlags, opaque: npxOpaque},
	{words: []string{"bunx"}, valueFlags: npxFlags, opaque: npxOpaque},
	{words: []string{"pnpm", "exec"}},
	{words: []string{"pnpm", "dlx"}},
	{words: []string{"uv", "run"}, valueFlags: uvFlags},
	{words: []string{"uvx"}, valueFlags: uvFlags},
	{words: []string{"dotnet", "tool", "run"}},
}

// unwrap drops a leading package runner and its options from the words of a
// command and returns what is left. It reports false when the command does not
// start with a runner, an option hides the command, or nothing is left.
func unwrap(words []string) ([]string, bool) {
	for _, r := range runners {
		if !hasPrefix(words, r.words) {
			continue
		}
		rest := words[len(r.words):]
		for len(rest) > 0 && strings.HasPrefix(rest[0], "-") {
			flag, _, _ := strings.Cut(rest[0], "=")
			switch {
			case r.opaque[flag]:
				return nil, false
			case rest[0] == "--":
				rest = rest[1:]
				return rest, len(rest) > 0
			case r.valueFlags[rest[0]] && len(rest) > 1:
				rest = rest[2:]
			default:
				rest = rest[1:]
			}
		}
		return rest, len(rest) > 0
	}
	return nil, false
}

func hasPrefix(words, prefix []string) bool {
	if len(words) < len(prefix) {
		return false
	}
	for i, w := range prefix {
		if words[i] != w {
			return false
		}
	}
	return true
}

// keys returns the forms a command is matched under: as written and, when it
// starts with a package runner, unwrapped. Both get the canonical rules.
func keys(cmd string) []string {
	out := []string{canonical(cmd)}
	if rest, ok := unwrap(strings.Fields(cmd)); ok {
		if k := canonical(strings.Join(rest, " ")); k != out[0] {
			out = append(out, k)
		}
	}
	return out
}
