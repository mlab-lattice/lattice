#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=${DIR}/../..
BUILD_CONTAINER_NAME="lattice-system-builder"

STATUS=$(docker inspect -f '{{.State.Running}}' ${BUILD_CONTAINER_NAME})
if [[ $? == "0" ]]; then
    if [[ ${STATUS} == "true" ]]; then
        echo "${BUILD_CONTAINER_NAME} container already running"
        exit 0
    else
        docker restart ${BUILD_CONTAINER_NAME}
        if [[ $? != "0" ]]; then
            echo "could not start ${BUILD_CONTAINER_NAME} container"
            exit 1
        fi
        exit 0
    fi
fi

set -e
docker run -d --name ${BUILD_CONTAINER_NAME} \
    -v ${ROOT_DIR}:/src -v /var/run/docker.sock:/var/run/docker.sock \
    -v ~/.ssh/id_rsa-github:/root/.ssh/id_rsa-github \
    -v ~/.config/gcloud:/root/.config/gcloud \
    lattice-build/bazel-build \
    sleep infinity
