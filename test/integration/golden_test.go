// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"testing"

	"github.com/jimmykurian/ci-parity/internal/cli"
	"github.com/jimmykurian/ci-parity/internal/testutil"
)

// TestGolden checks every fixture in every format against its goldens. Run
// `task test -- -update` to rewrite them.
func TestGolden(t *testing.T) {
	testutil.RunGolden(t, fixtures, testutil.Formats, cli.Run)
}
