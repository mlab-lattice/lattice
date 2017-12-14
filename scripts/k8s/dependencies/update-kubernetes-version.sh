#!/usr/bin/env bash

set -e
set -u

gsed -i -E "s/(\"tag\": \"kubernetes-).+(\",)/\1${KUBERNETES_VERSION}\2/g" ${LATTICE_ROOT}/bazel/go_repositories.bzl
