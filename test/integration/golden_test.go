// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jimmykurian/ci-parity/internal/cli"
	"github.com/jimmykurian/ci-parity/internal/testutil"
)

// goldenEntries returns one table entry per fixture and output format. A read
// error yields no entries; the "finds the fixtures" spec reports it.
func goldenEntries() []TableEntry {
	dirs, _ := testutil.FixtureDirs(fixtures)
	var entries []TableEntry
	for _, dir := range dirs {
		for _, format := range testutil.Formats {
			entries = append(entries, Entry(filepath.Base(dir)+", "+format, dir, format))
		}
	}
	return entries
}

// Run `task test -- -update` to rewrite the golden files from the current output.
var _ = Describe("ci-parity check output", Label("integration", "golden"), func() {
	It("finds the fixtures", func() {
		dirs, err := testutil.FixtureDirs(fixtures)
		Expect(err).NotTo(HaveOccurred())
		Expect(dirs).NotTo(BeEmpty())
	})

	DescribeTable("matches each fixture's golden file and exit code",
		func(dir, format string) {
			var stdout, stderr bytes.Buffer
			code := cli.Run([]string{"check", "-C", dir, "--format", format}, &stdout, &stderr)
			expected := filepath.Join(dir, "expected")
			testutil.CompareGolden(GinkgoT(), filepath.Join(expected, format+".golden"), stdout.Bytes(), stderr.String())
			testutil.CompareExitCode(GinkgoT(), filepath.Join(expected, "exitcode"), code, stderr.String())
		},
		goldenEntries(),
	)
})
