#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_ROOT=${DIR}/../..

minikube start \
    --kubernetes-version v1.10.1 \
    --bootstrapper kubeadm \
    --memory 4096 \
    --feature-gates=CustomResourceSubresources=true

${LATTICE_ROOT}/install/kubernetes/helm/install.sh ${@}
