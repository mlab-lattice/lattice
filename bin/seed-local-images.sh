#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMPONENT_BUILDER_DIR=${DIR}/../../component-builder
ENVOY_INTEGRATION_DIR=${DIR}/../../envoy-integration
WORKING_DIRECTORY=/tmp/local-lattice-dev

mkdir -p ${WORKING_DIRECTORY}

# Save kubernetes-integration images
cd ${DIR}/..
LATTICE_CONTROLLER_MANAGER_PATH=${WORKING_DIRECTORY}/lattice-controller-manager
BOOTSTRAP_KUBERNETES_PATH=${WORKING_DIRECTORY}/bootstrap-kubernetes
dest=${WORKING_DIRECTORY} make docker-save

# Save component-builder images
cd ${COMPONENT_BUILDER_DIR}
PULL_GIT_REPO_PATH=${WORKING_DIRECTORY}/pull-git-repo
BUILD_DOCKER_IMAGE_PATH=${WORKING_DIRECTORY}/build-docker-image
dest=${WORKING_DIRECTORY} make docker-save

# Save envoy-integration images
cd ${ENVOY_INTEGRATION_DIR}
PREPARE_ENVOY_PATH=${WORKING_DIRECTORY}/prepare-envoy
ENVOY_API_PATH=${WORKING_DIRECTORY}/kubernetes-per-node-rest
dest=${WORKING_DIRECTORY} make docker-save


# Load the images into minikube
eval $(minikube docker-env -p ${1})

images=(
    ${LATTICE_CONTROLLER_MANAGER_PATH}
    ${BOOTSTRAP_KUBERNETES_PATH}
    ${PULL_GIT_REPO_PATH}
    ${BUILD_DOCKER_IMAGE_PATH}
    ${PREPARE_ENVOY_PATH}
    ${ENVOY_API_PATH}
)

for i in ${images[@]}; do
    docker load -i ${i}
done
