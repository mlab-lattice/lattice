#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../..

${LATTICE_ROOT}/install/kubernetes/helm/install.sh --set cloudProvider.local.ip=$(minikube ip) ${@}
