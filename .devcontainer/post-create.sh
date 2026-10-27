#!/bin/bash
set -euo pipefail

# Runs once when the devcontainer is created: Go, the tools in mise.toml, and a
# first build.

cd "$(dirname "$0")/.."

# Go at the toolchain go.mod pins, falling back to its go line. mise.toml leaves
# Go out so go.mod stays the only pin; this installs it for the container only.
go_version=$(awk '$1 == "toolchain" { sub(/^go/, "", $2); print $2; exit }' go.mod)
go_version=${go_version:-$(awk '$1 == "go" { print $2; exit }' go.mod)}
mise use --global "go@${go_version}"
mise install

# New shells get the tools on PATH.
# shellcheck disable=SC2016 # written to .bashrc unexpanded
grep -q 'mise activate' ~/.bashrc || echo 'eval "$(mise activate bash)"' >> ~/.bashrc
eval "$(mise activate bash --shims)"

go mod download
task build

# The repo's git config may sign commits with a key file that only exists on
# the host. Say so now rather than on the first commit.
key=$(git config --get user.signingkey || true)
if [ -n "$key" ] && [ "${key#key::}" = "$key" ] && [ ! -f "${key/#\~/$HOME}" ]; then
  echo "note: user.signingkey is $key, which isn't in this container."
  echo "      Commit from the host, or see 'Signed commits' in CONTRIBUTING.md."
fi
