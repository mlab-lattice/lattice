#!/usr/bin/env bash

set -u

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../..

set -e
echo "pushing ${@}"
cd ${LATTICE_ROOT}
make docker.kubernetes.${@}.push

${DIR}/bounce-component.sh ${@}
