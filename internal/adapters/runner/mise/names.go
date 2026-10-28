// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package mise

import (
	"strings"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

// Command is how CI calls a mise task.
func (m *Mise) Command() string { return "mise run" }

// Internal is how P003 describes a hidden task.
func (m *Mise) Internal() string { return "is hidden" }

// otherDir are flags that run mise from another directory.
var otherDir = map[string]bool{"-C": true, "--cd": true}

// valueFlags take the next argument as their value.
var valueFlags = map[string]bool{
	"-j": true, "--jobs": true, "-E": true, "--env": true, "-t": true, "--tool": true,
	"--timeout": true, "-o": true, "--output": true, "--shell": true, "-s": true,
}

// Names returns the tasks a `mise run` call names: the first word after the
// flags, then the first word after each `:::`. The other words are task
// arguments. A call from another directory, or naming no task, returns none;
// a name that is not literal or holds a glob cannot be checked and is skipped.
func (m *Mise) Names(args []domain.Word) []domain.Word {
	var names []domain.Word
	expectTask, flags := true, true
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a.Text == ":::":
			expectTask, flags = true, true
		case !expectTask:
			// an argument of the current task
		case flags && a.Text == "--":
			flags = false
		case flags && (otherDir[a.Text] || strings.HasPrefix(a.Text, "--cd=") || (strings.HasPrefix(a.Text, "-C") && len(a.Text) > 2)):
			return nil
		case flags && valueFlags[a.Text]:
			i++
		case flags && strings.HasPrefix(a.Text, "-"):
		default:
			expectTask = false
			if a.Literal && !strings.ContainsAny(a.Text, "*?[") {
				names = append(names, a)
			}
		}
	}
	return names
}
