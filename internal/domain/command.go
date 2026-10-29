// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

// CommandKind classifies an extracted command.
type CommandKind int

const (
	// KindTool is a call to a command on the tools list.
	KindTool CommandKind = iota + 1
	// KindInvoker is a call to the task runner through a shell invoker.
	KindInvoker
)

// ExtractOptions configures command extraction.
type ExtractOptions struct {
	Tools       map[string]bool // tool basenames to report
	Transparent map[string]bool // wrapper basenames to look through
	Invokers    [][]string      // shell invokers as word sequences, e.g. {"go", "tool", "task"}
}

// Word is one argument of a command.
type Word struct {
	Text      string // literal value, or the source text when not literal
	Literal   bool
	Line, Col int // 1-based position within the script
}

// Command is one tool or invoker call found in a script.
type Command struct {
	Kind       CommandKind
	Name       string // tool basename, or the invoker as configured
	Normalized string // literal words from the command on, joined by single spaces
	Partial    bool   // a word from the command on isn't literal, so Normalized leaves it out
	Args       []Word // invoker arguments after the invoker words
	Line, Col  int    // 1-based position within the script
	Allowed    bool
	Reason     string
}

// Problem is a condition inside a script that prevents analysis (rule P005).
type Problem struct {
	Line, Col int
	Message   string
}

// Extraction holds everything extracted from one script, sorted by position.
type Extraction struct {
	Commands []Command
	Problems []Problem
}

// ParseError reports a script that does not parse. Line and Col are
// 1-based positions in the script, or 0 when the parser gave none.
type ParseError struct {
	Line, Col int
	Message   string
}

// Error returns the parser's message.
func (e *ParseError) Error() string {
	return e.Message
}
