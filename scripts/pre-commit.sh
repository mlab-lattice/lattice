#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../..
cd ${LATTICE_ROOT}

make gazelle

ERROR_MSG="Please run \"make format\" prior to committing"
[[ $(gofmt -l .) ]] && echo ${ERROR_MSG} && exit 1
[[ $(terraform fmt -list -write=false .) ]] && echo ${ERROR_MSG} && exit 1

exit 0
