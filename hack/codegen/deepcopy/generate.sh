#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../../..

deepcopy-gen --alsologtostderr \
    -h ${LATTICE_ROOT}/hack/codegen/go-header.txt \
    --input-dirs github.com/mlab-lattice/lattice/pkg/api/v1 \
    -O zz_generated.deepcopy

deepcopy-gen --alsologtostderr \
    -h ${LATTICE_ROOT}/hack/codegen/go-header.txt \
    --input-dirs github.com/mlab-lattice/lattice/pkg/definition/v1 \
    -O zz_generated.deepcopy
