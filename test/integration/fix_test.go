// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"path/filepath"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jimmykurian/ci-parity/internal/cli"
)

// fixFixture is the repository fix rewrites, with the workflow it should end up as.
const fixFixture = "../../testdata/fix/basic"

// The edits fix makes, in order. A dry run can't know the escaped value on line
// 28 will fail the editor's check, so it plans that one too; the real run skips
// it and counts it as a manual change.
const fixEdits = ".github/workflows/ci.yml:9:14 'pnpm install --frozen-lockfile' -> 'task install'\n" +
	".github/workflows/ci.yml:11:14 'pnpm test' -> 'task test'\n" +
	".github/workflows/ci.yml:13:19 'pnpm lint' -> 'task lint'\n" +
	".github/workflows/ci.yml:15:43 'pnpm build' -> 'task build'\n" +
	".github/workflows/ci.yml:17:28 'pnpm test' -> 'task test'\n" +
	".github/workflows/ci.yml:19:15 'pnpm test' -> 'task test'\n" +
	".github/workflows/ci.yml:21:15 'pnpm lint' -> 'task lint'\n" +
	".github/workflows/ci.yml:24:11 'pnpm test' -> 'task test'\n"

var _ = Describe("ci-parity fix", Label("integration"), func() {
	workflow := filepath.Join(".github", "workflows", "ci.yml")
	var repo string

	BeforeEach(func() {
		repo = copyDir(filepath.Join(fixFixture, "repo"))
	})

	It("lists 9 planned edits on a dry run and writes nothing", func() {
		before := readFile(filepath.Join(repo, workflow))
		r := runCLI("fix", "--dry-run", "-C", repo)
		Expect(r.code).To(Equal(cli.ExitOK), "stderr: %s", r.stderr)
		Expect(r.stdout).To(Equal(fixEdits +
			".github/workflows/ci.yml:28:15 'pnpm\ttest' -> 'task test'\n" +
			"would fix 9 steps in 1 file; 7 inline-tool findings need a manual change\n"))
		Expect(readFile(filepath.Join(repo, workflow))).To(Equal(before))
	})

	It("rewrites the 8 steps it can, matching the expected workflow", func() {
		r := runCLI("fix", "-C", repo)
		Expect(r.code).To(Equal(cli.ExitOK), "stderr: %s", r.stderr)
		Expect(r.stdout).To(Equal(fixEdits + "fixed 8 steps in 1 file; 8 inline-tool findings need a manual change\n"))
		want := readFile(filepath.Join(fixFixture, "expected", workflow))
		got := readFile(filepath.Join(repo, workflow))
		Expect(cmp.Diff(want, got)).To(BeEmpty(), "fixed workflow (-want +got)")
	})

	It("changes nothing on a second run", func() {
		Expect(runCLI("fix", "-C", repo).code).To(Equal(cli.ExitOK))
		r := runCLI("fix", "-C", repo)
		Expect(r.stdout).To(Equal("fixed 0 steps in 0 files; 8 inline-tool findings need a manual change\n"))
	})

	It("rewrites only steps whose working directory matches the task's", func() {
		wd := copyDir(filepath.Join(fixtures, "working-directory"))
		r := runCLI("fix", "--dry-run", "-C", wd)
		Expect(r.code).To(Equal(cli.ExitOK), "stderr: %s", r.stderr)
		Expect(r.stdout).To(Equal(".github/workflows/ci.yml:8:14 'pnpm test' -> 'task test'\n" +
			".github/workflows/ci.yml:10:14 'pnpm fmt' -> 'task tools:fmt'\n" +
			".github/workflows/ci.yml:15:14 'pnpm lint' -> 'task lint-api'\n" +
			".github/workflows/ci.yml:31:14 'pnpm build' -> 'task web:build'\n" +
			".github/workflows/ci.yml:34:14 'pnpm test' -> 'task web:test'\n" +
			".github/workflows/ci.yml:37:14 'pnpm e2e' -> 'task web:e2e'\n" +
			"would fix 6 steps in 1 file; 6 inline-tool findings need a manual change\n"))
	})

	It("suggests running baseline once the counts fell", func() {
		Expect(runCLI("baseline", "-C", repo).code).To(Equal(cli.ExitOK))
		r := runCLI("fix", "-C", repo)
		Expect(r.stdout).To(ContainSubstring("run `ci-parity baseline` to lock in the lower counts"))
	})
})
