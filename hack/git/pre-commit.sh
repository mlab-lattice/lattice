#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../..
cd ${LATTICE_ROOT}

echo "Checking BUILD files and formatting..."

GAZELLE_ERROR_MSG="Please run \"make gazelle\" and add the generated BUILD.bazel files prior to committing"
[[ $(bazel run -- //:gazelle -mode diff 2>/dev/null) ]] && echo ${GAZELLE_ERROR_MSG} && exit 1

FMT_ERROR_MSG="Please run \"make format\" and add the fixed files prior to committing"
if ! gofmt -l .; then echo ${FMT_ERROR_MSG} && exit 1; fi
[[ $(gofmt -l .) ]] && echo ${FMT_ERROR_MSG} && exit 1
[[ $(terraform fmt -list -write=false .) ]] && echo ${FMT_ERROR_MSG} && exit 1

exit 0
