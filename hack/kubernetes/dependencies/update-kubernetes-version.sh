#!/usr/bin/env bash

set -e
set -u

gsed -i -E "s/(\"tag\": \"kubernetes-).+(\",)/\1${KUBERNETES_VERSION}\2/g" ${LATTICE_ROOT}/bazel/go/dependencies.bzl

dependencies=(
    "k8s.io/api"
    "k8s.io/apimachinery"
    "k8s.io/apiextensions-apiserver"
    "k8s.io/client-go"
)

for d in ${dependencies[@]}; do
    cd ${GOPATH}/src/${d}
    git fetch --tags origin
    git checkout kubernetes-${KUBERNETES_VERSION}
done
