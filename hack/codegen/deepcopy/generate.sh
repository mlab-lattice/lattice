#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../../..

declare -a packages=(
    "api/v1"
    "definition/v1"
    "definition/tree"
    "definition/resolver"
    "definition/resolver/template"
)

## now loop through the above array
for p in "${packages[@]}"
do
    deepcopy-gen --alsologtostderr \
        -h ${LATTICE_ROOT}/hack/codegen/go-header.txt \
        --input-dirs github.com/mlab-lattice/lattice/pkg/${p} \
        -O zz_generated.deepcopy
done

