#!/usr/bin/env sh

set -o errexit
set -o pipefail
set -o nounset
set -o xtrace

cd lattice-repo
TAG_NAME=$(git for-each-ref --format='%(refname:short)' refs/heads)
cd ..

mkdir -p ./docs-html/$TAG_NAME

echo "Building DOCS: ./docs-html/$TAG_NAME"

./docgen-binary/docgen --output-docs ./docs-html/latticectl-reference-$TAG_NAME.md --input-docs ./lattice-repo/docs/cli

#./docgen-binary/docgen --output-docs ./docs-html/latticectl-reference.md --input-docs ./lattice-repo/docs/cli
