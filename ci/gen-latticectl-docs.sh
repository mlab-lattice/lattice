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

ls -la tar-temp

# tar -xvf latticectl-docs-markdown-bucket/latticectl-docs.tar.gz -C tar-temp

./docgen-binary/docgen --output-docs ./tar-temp/latticectl-reference-$TAG_NAME.md --input-docs ./lattice-repo/docs/cli

ls -la tar-temp

tar cvzf ./latticectl-docs-markdown-bucket/latticectl-docs.tar.gz ./tar-temp

#./docgen-binary/docgen --output-docs ./docs-html/latticectl-reference.md --input-docs ./lattice-repo/docs/cli
