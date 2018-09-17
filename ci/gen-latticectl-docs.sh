#!/usr/bin/env sh

set -o errexit
set -o pipefail
set -o nounset
set -o xtrace

cd lattice-repo
TAG_NAME=$(git for-each-ref --format='%(refname:short)' refs/heads)
cd ..

echo "Building DOCS: latticectl-referene-$TAG_NAME.md"

mkdir tar-temp

./docgen-binary/docgen --output-docs ./tar-temp/latticectl-reference-$TAG_NAME.md --input-docs ./lattice-repo/docs/cli

# We use tarballs in case we want to add multiple files in the future
tar cvzf ./tarball/latticectl-docs-$TAG_NAME.tar.gz ./tar-temp/*
