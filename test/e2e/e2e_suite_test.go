// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

// Package e2e builds the ci-parity binary once and runs it as a separate
// process, checking what only a real process shows: the version stamped in with
// -ldflags, exit codes, the split between stdout and stderr, the working
// directory standing in for -C, and files written to disk. The specs are
// written with Ginkgo and Gomega; gexec builds and runs the binary.
package e2e

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// stamp is the build metadata the binary is built with, as `version` prints it.
const stamp = "ci-parity e2e (abc1234, 2026-01-02T03:04:05Z)"

// binary is the path of the ci-parity binary BeforeSuite builds.
var binary string

// TestE2E runs every spec in the package.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "End-to-end")
}

var _ = BeforeSuite(func() {
	pkg := "github.com/jimmykurian/ci-parity/internal/cli"
	ldflags := fmt.Sprintf("-X %[1]s.version=e2e -X %[1]s.commit=abc1234 -X %[1]s.date=2026-01-02T03:04:05Z", pkg)
	var err error
	binary, err = gexec.Build("github.com/jimmykurian/ci-parity/cmd/ci-parity", "-ldflags", ldflags)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
