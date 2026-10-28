// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Command ci-parity checks that GitHub Actions workflows run project commands
// through the repository's task runner instead of calling tools inline.
package main

import (
	"os"

	"github.com/jimmykurian/ci-parity/internal/cli"
)

// main runs the command line and exits with its code.
func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
