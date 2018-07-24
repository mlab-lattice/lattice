#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
set -o xtrace

TAG_NAME=$(cd lattice-repo; git rev-parse --short HEAD)

mkdir -p ./docs-html/$(TAG_NAME)

./docgen-binary/docgen --output-docs ./docs-html/$(TAG_NAME)/latticectl-reference.md --input-docs ./lattice-repo/docs/cli
