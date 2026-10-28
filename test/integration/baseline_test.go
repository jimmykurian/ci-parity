// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jimmykurian/ci-parity/internal/cli"
)

var _ = Describe("ci-parity baseline", Label("integration"), func() {
	Context("when the inline-tool counts rose", func() {
		var dir, path, before string

		BeforeEach(func() {
			dir = copyDir(fixture("baseline-regression"))
			path = filepath.Join(dir, ".ci-parity-baseline.json")
			before = readFile(path)
		})

		It("refuses, listing the workflow, and leaves the file unchanged", func() {
			r := runCLI("baseline", "-C", dir)
			Expect(r.code).To(Equal(cli.ExitFindings))
			Expect(r.stderr).To(ContainSubstring(".github/workflows/ci.yml  1 -> 3"))
			Expect(readFile(path)).To(Equal(before))
		})

		It("writes the higher counts with --allow-regression", func() {
			r := runCLI("baseline", "-C", dir, "--allow-regression")
			Expect(r.code).To(Equal(cli.ExitOK), "stderr: %s", r.stderr)
			Expect(readFile(path)).To(Equal("{\n  \"version\": 1,\n  \"workflows\": {\n    \".github/workflows/ci.yml\": 3\n  }\n}\n"))
			Expect(r.stdout).To(ContainSubstring("1 workflow, 3 inline-tool findings"))
		})
	})

	Context("when there is no baseline file", func() {
		It("creates one, and check then tolerates every finding", func() {
			dir := copyDir(fixture("baseline-locked"))
			Expect(os.Remove(filepath.Join(dir, ".ci-parity-baseline.json"))).To(Succeed())

			r := runCLI("baseline", "-C", dir)
			Expect(r.code).To(Equal(cli.ExitOK), "stderr: %s", r.stderr)
			Expect(readFile(filepath.Join(dir, ".ci-parity-baseline.json"))).To(Equal(
				"{\n  \"version\": 1,\n  \"workflows\": {\n" +
					"    \".github/workflows/ci.yml\": 3,\n" +
					"    \".github/workflows/nightly.yml\": 1,\n" +
					"    \".github/workflows/release.yml\": 1\n  }\n}\n"))

			r = runCLI("check", "-C", dir)
			Expect(r.code).To(Equal(cli.ExitOK))
			Expect(r.stdout).To(ContainSubstring("5 tolerated"))
		})
	})
})

var _ = Describe("ci-parity check with a baseline", Label("integration"), func() {
	DescribeTable("applies or ignores the baseline",
		func(args []string, wantCode int, want string) {
			r := runCLI(append([]string{"check", "-C", fixture("baseline-within")}, args...)...)
			Expect(r.code).To(Equal(wantCode), "stdout: %s\nstderr: %s", r.stdout, r.stderr)
			Expect(r.stdout).To(ContainSubstring(want))
		},
		Entry("tolerates findings up to each count", nil, cli.ExitOK, "3 tolerated"),
		Entry("ignores it with --strict", []string{"--strict"}, cli.ExitFindings, "3 errors, 0 warnings, 0 tolerated"),
		Entry("shows tolerated findings with --verbose", []string{"--verbose"}, cli.ExitOK, "runs 'pnpm install' directly [tolerated]"),
		Entry("is strict when an explicit --baseline is missing", []string{"--baseline", "none.json"}, cli.ExitFindings, "3 errors"),
	)
})
