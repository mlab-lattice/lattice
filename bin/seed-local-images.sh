#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMPONENT_BUILDER_DIR=${DIR}/../../component-builder
ENVOY_INTEGRATION_DIR=${DIR}/../../envoy-integration

eval $(minikube docker-env -p ${1})

cd ${COMPONENT_BUILDER_DIR}
bazel run //docker:pull-git-repo
bazel run //docker:build-docker-image

cd ${ENVOY_INTEGRATION_DIR}
make docker-build-prepare-envoy
