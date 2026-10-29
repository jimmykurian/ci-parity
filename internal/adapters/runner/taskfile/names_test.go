// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package taskfile

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/jimmykurian/ci-parity/internal/domain"
)

func TestTaskNames(t *testing.T) {
	// Arrange
	tests := []struct {
		args string
		want []string
	}{
		{"build lint", []string{"build", "lint"}},
		{"build VERSION=1.2.3", []string{"build"}},
		{"-C 2 build", []string{"build"}},
		{"--concurrency=2 test", []string{"test"}},
		{"-o group --output prefixed build", []string{"build"}},
		{"--interval 5s lint", []string{"lint"}},
		{"--silent --force test", []string{"test"}},
		{"build -- --verbose extra", []string{"build"}},
		{"--list", nil},
		{"-t other.yml migrate", nil},
		{"--taskfile=other.yml migrate", nil},
		{"-d web build", nil},
		{"--dir=web build", nil},
		{"-dweb build", nil},
		{"--global setup", nil},
		{"-g setup", nil},
		{"-o group --output-group-begin ::group::{{.TASK}} --output-group-end ::endgroup:: ci", []string{"ci"}},
		{"--sort alphanumeric --list-all", nil},
		{"--temp-dir .cache/task --timeout 30s --expiry 1h build", []string{"build"}},
		{"--cacert ca.pem --cert client.pem --cert-key client.key build", []string{"build"}},
		{"--remote-cache-dir .cache --trusted-hosts example.com build", []string{"build"}},
		{"--completion bash", nil},
		{"--completion=zsh", nil},
	}
	for _, tt := range tests {
		// Arrange
		var args []domain.Word
		for _, f := range strings.Fields(tt.args) {
			args = append(args, domain.Word{Text: f, Literal: true})
		}

		// Act
		names := (&Taskfile{}).Names(args)

		// Assert
		var got []string
		for _, w := range names {
			got = append(got, w.Text)
		}
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("(&Taskfile{}).Names(%q) (-want +got):\n%s", tt.args, diff)
		}
	}
}

func TestTaskNamesSkipsNonLiteral(t *testing.T) {
	// Arrange
	dynamic := []domain.Word{{Text: "${{ inputs.target }}"}, {Text: "build", Literal: true}}

	// Act
	got := (&Taskfile{}).Names(dynamic)

	// Assert
	if len(got) != 1 || got[0].Text != "build" {
		t.Errorf("non-literal names are skipped, got %v", got)
	}
}
