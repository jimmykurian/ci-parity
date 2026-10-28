// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jimmykurian/ci-parity/internal/cli"
)

// fixtures holds the golden fixture repositories, relative to this package.
const fixtures = "../../testdata/fixtures"

// fixture returns the path of the named golden fixture.
func fixture(name string) string {
	return filepath.Join(fixtures, name)
}

// copyDir copies the fixture repo at src into a new temporary directory, so a
// spec can write to it.
func copyDir(src string) string {
	dst := GinkgoT().TempDir()
	Expect(os.CopyFS(dst, os.DirFS(src))).To(Succeed())
	return dst
}

// result is what one in-process run of ci-parity produced.
type result struct {
	code           int
	stdout, stderr string
}

// runCLI runs ci-parity with args in-process.
func runCLI(args ...string) result {
	var stdout, stderr bytes.Buffer
	code := cli.Run(args, &stdout, &stderr)
	return result{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

// readFile returns the contents of path, failing the spec when it can't be read.
func readFile(path string) string {
	GinkgoHelper()
	data, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}
