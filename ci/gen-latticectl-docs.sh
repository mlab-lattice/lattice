#!/usr/bin/env sh

set -o errexit
set -o pipefail
set -o nounset
set -o xtrace

cd lattice-repo
TAG_NAME=$(git rev-parse --abbrev-ref HEAD)
cd ..

mkdir -p ./docs-html/$TAG_NAME

echo "Building DOCS: ./docs-html/$TAG_NAME"

./docgen-binary/docgen --output-docs ./docs-html/$TAG_NAME/latticectl-reference.md --input-docs ./lattice-repo/docs/cli

#./docgen-binary/docgen --output-docs ./docs-html/latticectl-reference.md --input-docs ./lattice-repo/docs/cli
