// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

// fixtures holds the golden fixture repositories, relative to this package.
const fixtures = "../../testdata/fixtures"

// fixture returns the path of the named golden fixture.
func fixture(name string) string {
	return filepath.Join(fixtures, name)
}

// start runs the binary with args in dir and returns its session.
func start(dir string, args ...string) *gexec.Session {
	GinkgoHelper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

// finish waits for session to exit with code.
func finish(session *gexec.Session, code int) {
	GinkgoHelper()
	Eventually(session).WithTimeout(gexecTimeout).Should(gexec.Exit(code),
		"stdout: %s\nstderr: %s", session.Out.Contents(), session.Err.Contents())
}

// gexecTimeout bounds one run of the binary; a check takes well under a second.
const gexecTimeout = 10 * time.Second

// golden returns a fixture's expected output for format.
func golden(name, format string) string {
	GinkgoHelper()
	data, err := os.ReadFile(filepath.Join(fixture(name), "expected", format+".golden"))
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

var _ = Describe("ci-parity version", Label("e2e"), func() {
	It("prints the version stamped in at build time", func() {
		s := start(".", "version")
		finish(s, 0)
		Expect(string(s.Out.Contents())).To(Equal(stamp + "\n"))
		Expect(s.Err.Contents()).To(BeEmpty())
	})
})

var _ = Describe("exit codes and output streams", Label("e2e"), func() {
	// wantStdout and wantStderr are substrings; "" means that stream stays empty.
	DescribeTable("exits with the documented code and writes to the right stream",
		func(args []string, code int, wantStdout, wantStderr string) {
			s := start(".", args...)
			finish(s, code)
			for _, stream := range []struct {
				buffer *gbytes.Buffer
				want   string
			}{{s.Out, wantStdout}, {s.Err, wantStderr}} {
				if stream.want == "" {
					Expect(stream.buffer.Contents()).To(BeEmpty())
				} else {
					Expect(stream.buffer).To(gbytes.Say(regexp.QuoteMeta(stream.want)))
				}
			}
		},
		Entry("0 for a clean repository", []string{"check", "-C", fixture("clean")}, 0, "0 errors", ""),
		Entry("1 for error findings", []string{"check", "-C", fixture("inline-tools")}, 1, "P001 inline-tool", ""),
		Entry("0 for --help, on stdout", []string{"--help"}, 0, "Usage:", ""),
		Entry("2 with no command, usage on stderr", nil, 2, "", "Usage:"),
		Entry("2 for an unknown command", []string{"lint"}, 2, "", `unknown command "lint"`),
		Entry("2 for a missing -C directory", []string{"check", "-C", "absent"}, 2, "", "-C absent: not a directory"),
	)
})

var _ = Describe("run from the repository directory", Label("e2e"), func() {
	DescribeTable("checks the working directory when -C is left out",
		func(format string) {
			s := start(fixture("inline-tools"), "check", "--format", format)
			finish(s, 1)
			Expect(string(s.Out.Contents())).To(Equal(golden("inline-tools", format)))
		},
		Entry("text", "text"),
		Entry("json", "json"),
	)
})

var _ = Describe("ci-parity fix", Label("e2e"), func() {
	It("rewrites the copied workflow on disk", func() {
		src := "../../testdata/fix/basic"
		repo := GinkgoT().TempDir()
		Expect(os.CopyFS(repo, os.DirFS(filepath.Join(src, "repo")))).To(Succeed())

		s := start(repo, "fix")
		finish(s, 0)
		Expect(s.Out).To(gbytes.Say("fixed 8 steps in 1 file"))

		workflow := filepath.Join(".github", "workflows", "ci.yml")
		got, err := os.ReadFile(filepath.Join(repo, workflow))
		Expect(err).NotTo(HaveOccurred())
		want, err := os.ReadFile(filepath.Join(src, "expected", workflow))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(got)).To(Equal(string(want)))
	})
})
