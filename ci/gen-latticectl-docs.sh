#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
set -o xtrace

./docgen-binary/docgen --output-docs ./docs-html/latticectl-reference.md --input-docs ./lattice-repo/docs/cli/
