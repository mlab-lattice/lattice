#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMPONENT_BUILDER_DIR=${DIR}/../../component-builder

cd ${COMPONENT_BUILDER_DIR}

eval $(minikube docker-env)
bazel run //docker:pull-git-repo
bazel run //docker:build-docker-image
