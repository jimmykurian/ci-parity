// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package rules

import (
	"strings"
)

// suggester proposes tasks for inline commands and unknown task names. Internal
// tasks are never proposed: task refuses to run them from the command line.
type suggester struct {
	invoke    string                 // how CI calls a task, such as "task"
	reaches   func(dir string) bool  // whether a call from dir reads the runner's file
	byCommand map[string][]candidate // canonical command -> tasks that run it
	names     []string               // invocable, non-internal names
}

// candidate is a task that runs a command, and the directory it runs in.
type candidate struct {
	name     string
	dir      string
	dirKnown bool
}

// newSuggester indexes the runner's non-internal tasks by the commands they run.
func newSuggester(r TaskSource) *suggester {
	s := &suggester{byCommand: map[string][]candidate{}, invoke: r.Command(), reaches: r.Reaches}
	for _, name := range r.Tasks() {
		if r.IsInternal(name) {
			continue
		}
		s.names = append(s.names, name)
		dir, known := r.Dir(name)
		for _, cmd := range r.Commands(name) {
			for _, key := range keys(cmd) {
				s.byCommand[key] = append(s.byCommand[key], candidate{name: name, dir: dir, dirKnown: known})
			}
		}
	}
	return s
}

// command returns "<invoke> <name>" for a task whose command equals the inline
// command, as written or with a package runner unwrapped on either side, and
// that runs in dir, the step's working directory; or "". Calling the runner
// from dir must read the same file, so the name resolves. A task whose
// directory isn't known never matches, nor does any task when dirKnown is
// false. Among several matching tasks the better name wins.
func (s *suggester) command(normalized, dir string, dirKnown bool) string {
	if !dirKnown || !s.reaches(dir) {
		return ""
	}
	best := ""
	for _, key := range keys(normalized) {
		for _, c := range s.byCommand[key] {
			if c.dirKnown && c.dir == dir && (best == "" || better(c.name, best)) {
				best = c.name
			}
		}
	}
	if best == "" {
		return ""
	}
	return s.invoke + " " + best
}

// closest returns the task name nearest to an unknown one. Names that start
// with the unknown name come first, shortest first; otherwise the name with
// the lowest Levenshtein distance, which must be at most 3 and less than the
// unknown name's length so short names do not match unrelated ones. It
// returns "" when no name qualifies.
func (s *suggester) closest(unknown string) string {
	prefix, near, nearDist := "", "", -1
	for _, name := range s.names {
		if strings.HasPrefix(name, unknown) {
			if prefix == "" || better(name, prefix) {
				prefix = name
			}
			continue
		}
		d := levenshtein(unknown, name)
		if d > 3 || d >= len([]rune(unknown)) {
			continue
		}
		if nearDist < 0 || d < nearDist || (d == nearDist && better(name, near)) {
			near, nearDist = name, d
		}
	}
	if prefix != "" {
		return prefix
	}
	return near
}

// better prefers the shorter name, then the alphabetically first.
func better(a, b string) bool {
	if len(a) != len(b) {
		return len(a) < len(b)
	}
	return a < b
}

// canonical collapses whitespace and treats `pnpm run X` as `pnpm X` and
// `yarn run X` as `yarn X`.
func canonical(cmd string) string {
	f := strings.Fields(cmd)
	if len(f) >= 3 && (f[0] == "pnpm" || f[0] == "yarn") && f[1] == "run" {
		f = append(f[:1], f[2:]...)
	}
	return strings.Join(f, " ")
}

// levenshtein returns the edit distance between a and b, counted in runes.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}
