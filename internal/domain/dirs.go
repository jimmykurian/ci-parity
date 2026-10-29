// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"path"
	"path/filepath"
	"strings"
)

// RepoDir returns the repo-relative directory a working-directory or task dir
// names, cleaned and with forward slashes, "" for the repository root. It
// returns false when that can't be known statically: an expression, a
// variable, a template, an absolute path or one that leaves the repository.
func RepoDir(dir string) (string, bool) {
	if strings.ContainsAny(dir, "$`{}~") || strings.HasPrefix(dir, "/") || strings.HasPrefix(dir, `\`) ||
		(len(dir) > 1 && dir[1] == ':') {
		return "", false
	}
	clean := path.Clean(strings.ReplaceAll(dir, `\`, "/"))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", false
	}
	if clean == "." {
		return "", true
	}
	return clean, true
}

// Reaches reports whether a runner started in the directory start reads file,
// when it looks for its file by walking up from where it starts: start is at
// or below file's directory, and found, the runner's own lookup, finds no file
// in start or a directory between them. found does the reading, so this stays
// free of I/O.
func Reaches(file, start string, found func(dir string) (string, bool)) bool {
	top := filepath.Dir(file)
	rel, err := filepath.Rel(top, start)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	for dir := filepath.Clean(start); dir != top; dir = filepath.Dir(dir) {
		if _, ok := found(dir); ok {
			return false
		}
	}
	return true
}
