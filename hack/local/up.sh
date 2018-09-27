#!/usr/bin/env bash

set -e
set -u

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../..

minikube start \
    --kubernetes-version v1.10.1 \
    --bootstrapper kubeadm \
    --memory 4096 \
    --vm-driver ${VM_DRIVER} \
    --feature-gates=CustomResourceSubresources=true

set +u
${LATTICE_ROOT}/install/kubernetes/helm/install.sh --set cloudProvider.local.ip=$(minikube ip) ${@}
