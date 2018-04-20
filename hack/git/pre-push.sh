#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../..
cd ${LATTICE_ROOT}

[[ $(git rev-parse --abbrev-ref HEAD) != "master" ]] && exit 0
echo "Attempting to push to master, running pre-push checks..."

echo "Running \"make build\"..."
make build > /dev/null 2>&1
if [[ $? -ne 0 ]]; then echo "\"make build\" failed" && exit 1; fi

set -e
echo "Running \"make test\"..."
make test
