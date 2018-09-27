#!/usr/bin/env sh

set -o errexit
set -o pipefail
set -o nounset
set -o xtrace

cd lattice-repo
TAG_NAME=$(git for-each-ref --format='%(refname:short)' refs/heads)

echo "Building DOCS: latticectl-reference-$TAG_NAME.md"

make docgen.latticectl.tar
cp bazel-bin/cmd/latticectl/docs-tar.tar ../tarball/latticectl-docs-${TAG_NAME}.tar.gz

