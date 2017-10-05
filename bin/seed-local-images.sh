#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMPONENT_BUILDER_DIR=${DIR}/../../component-builder
ENVOY_INTEGRATION_DIR=${DIR}/../../envoy-integration
WORKING_DIRECTORY=/tmp/lattice-kube-bootstrap-build

mkdir -p ${WORKING_DIRECTORY}

# Build component-builder images
cd ${COMPONENT_BUILDER_DIR}
PULL_GIT_REPO_PATH=${WORKING_DIRECTORY}/pull-git-repo.tar
dest=${PULL_GIT_REPO_PATH} make docker-save-pull-git-repo

BUILD_DOCKER_IMAGE_PATH=${WORKING_DIRECTORY}/build-docker-image.tar
dest=${BUILD_DOCKER_IMAGE_PATH} make docker-save-build-docker-image

# Build envoy-integration images
cd ${ENVOY_INTEGRATION_DIR}
PREPARE_ENVOY_PATH=${WORKING_DIRECTORY}/prepare-envoy.tar
dest=${PREPARE_ENVOY_PATH} make docker-save-prepare-envoy

#ENVOY_API_PATH=${WORKING_DIRECTORY}/envoy-api.tar
#dest=${ENVOY_API_PATH} make docker-save-kubernetes-per-node-rest

# Load the images into minikube
eval $(minikube docker-env -p ${1})

images=(
    ${PULL_GIT_REPO_PATH}
    ${BUILD_DOCKER_IMAGE_PATH}
    ${PREPARE_ENVOY_PATH}
#    ${ENVOY_API_PATH}
)

for i in ${images[@]}; do
    docker load -i ${i}
done
