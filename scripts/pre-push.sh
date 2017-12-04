#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../..
cd ${LATTICE_ROOT}

make build
make test
