// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

import "strings"

// MatchPath reports whether the repo-relative, forward-slash path name matches
// pattern. `*` matches any characters except `/`, `?` matches one character
// except `/`, and a whole `**` segment matches zero or more path segments.
func MatchPath(pattern, name string) bool {
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchSegments(pats, segs []string) bool {
	for len(pats) > 0 {
		if pats[0] == "**" {
			for i := 0; i <= len(segs); i++ {
				if matchSegments(pats[1:], segs[i:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 || !wildcard(pats[0], segs[0]) {
			return false
		}
		pats, segs = pats[1:], segs[1:]
	}
	return len(segs) == 0
}

// MatchCommand reports whether a normalized command matches pattern. `*`
// matches any run of characters, including `/` and spaces, and `?` matches
// exactly one character. Matching is case-sensitive.
func MatchCommand(pattern, command string) bool {
	return wildcard(pattern, command)
}

// wildcard matches s against pattern where `*` matches any run of characters
// and `?` matches one character. Path callers split on `/` first, so neither
// wildcard crosses a segment there.
func wildcard(pattern, s string) bool {
	p, n := []rune(pattern), []rune(s)
	pi, ni := 0, 0
	star, mark := -1, 0
	for ni < len(n) {
		switch {
		case pi < len(p) && (p[pi] == '?' || p[pi] == n[ni]):
			pi++
			ni++
		case pi < len(p) && p[pi] == '*':
			star, mark = pi, ni
			pi++
		case star >= 0:
			pi = star + 1
			mark++
			ni = mark
		default:
			return false
		}
	}
	for pi < len(p) && p[pi] == '*' {
		pi++
	}
	return pi == len(p)
}

// StaticPrefix returns the leading segments of a path pattern that contain no
// wildcard, so discovery can start walking there instead of at the repo root.
func StaticPrefix(pattern string) string {
	segs := strings.Split(pattern, "/")
	var prefix []string
	for _, s := range segs[:len(segs)-1] {
		if strings.ContainsAny(s, "*?") {
			break
		}
		prefix = append(prefix, s)
	}
	return strings.Join(prefix, "/")
}
