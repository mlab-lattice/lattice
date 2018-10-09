#!/usr/bin/env bash

set -e
set -u

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

minikube start \
    --kubernetes-version v1.10.7 \
    --bootstrapper kubeadm \
    --memory 4096 \
    --vm-driver ${VM_DRIVER} \
    --feature-gates=CustomResourceSubresources=true

${DIR}/install.sh
