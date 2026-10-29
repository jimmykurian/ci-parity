// Copyright (c) 2026 Jimmy Kurian. Licensed under Apache-2.0; see LICENSE.
// SPDX-License-Identifier: Apache-2.0

package domain

import "testing"

func TestBuiltinAllow(t *testing.T) {
	// Arrange
	tests := map[string]string{
		// version and info
		"go version":                ReasonVersion,
		"node --version":            ReasonVersion,
		"cargo -V":                  ReasonVersion,
		"kubectl version --client":  ReasonVersion,
		"go env GOVERSION":          ReasonVersion,
		"npm version patch":         "", // bumps the package version
		"cargo -V extra":            "",
		"python -v script.py":       "", // -v is verbose, not version
		"terraform version upgrade": "",
		// pinned installs
		"go install golang.org/x/vuln/cmd/govulncheck@v1.1.4": ReasonInstall,
		"go install -v a.dev/x@v1 b.dev/y@v2":                 ReasonInstall,
		"go install a.dev/x@latest":                           "",
		"go install a.dev/x@v1 b.dev/y":                       "", // every package pinned
		"go install ./cmd/widget":                             "",
		"cargo install --locked mdbook@0.4.52":                ReasonInstall,
		"cargo install mdbook --version 0.4.52":               ReasonInstall,
		"cargo install mdbook --version=0.4.52":               ReasonInstall,
		"cargo install mdbook":                                "",
		"cargo install --path .":                              "",
		"cargo install --git https://example.com/x.git x@1":   "",
		"npm install --global corepack@0.34.0":                ReasonInstall,
		"npm i -g @acme/cli@2.1.0":                            ReasonInstall,
		"npm i -g @acme/cli":                                  "", // a scope's @ is not a version
		"npm install corepack@0.34.0":                         "", // not global
		"npm install -g corepack@latest":                      "",
		"pnpm add -g turbo@2.5.0":                             ReasonInstall,
		"pip install poetry==2.2.1":                           ReasonInstall,
		"pip3 install --user poetry==2.2.1 tox==4.0":          ReasonInstall,
		"python3 -m pip install poetry==2.2.1":                ReasonInstall,
		"pip install poetry":                                  "",
		"pip install -r requirements.txt":                     "",
		"pip install -e .":                                    "",
		"pip install --requirement=requirements.txt x==1":     "",
		"python -m pip install --upgrade pip":                 "",
		// auth
		"docker login ghcr.io -u acme --password-stdin": ReasonAuth,
		"docker logout":                      ReasonAuth,
		"aws configure set region eu-west-1": ReasonAuth,
		"az login --identity":                ReasonAuth,
		"az account set --subscription x":    ReasonAuth,
		"az account show":                    "",
		"gcloud auth configure-docker":       ReasonAuth,
		"gcloud config set project acme":     ReasonAuth,
		"docker push ghcr.io/acme/widget":    "",
		// inline scripts
		"python3 -c print(1)":      ReasonScript,
		"python -":                 ReasonScript,
		"python3":                  ReasonScript, // a heredoc on standard input
		"node -e console.log(1)":   ReasonScript,
		"node --eval x":            ReasonScript,
		"node":                     ReasonScript,
		"python3 scripts/build.py": "",
		"node build.js":            "",
		"":                         "",
	}
	for command, want := range tests {
		// Act
		got, ok := BuiltinAllow(command)

		// Assert
		if got != want || ok != (want != "") {
			t.Errorf("BuiltinAllow(%q) = %q, %v; want %q", command, got, ok, want)
		}
	}
}
