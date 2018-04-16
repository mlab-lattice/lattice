#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_CONFIG_ROOT=${DIR}/../../.lattice

DOCKER_CONFIG_FILE_NAME=docker.json
DOCKER_CONFIG_FILE_PATH=${LATTICE_CONFIG_ROOT}/${DOCKER_CONFIG_FILE_NAME}
if [ ! -f ${DOCKER_CONFIG_FILE_PATH} ]; then
    echo "${DOCKER_CONFIG_FILE_NAME} does not exist"; exit 1
fi

command -v jq >/dev/null 2>&1 || { echo "jq not installed"; exit 1; }

CONFIG=$(cat ${DOCKER_CONFIG_FILE_PATH})
REGISTRY=$(echo ${CONFIG} | jq -eMr '.registry') || { echo "docker registry not set"; exit 1; }
CHANNEL=$(echo ${CONFIG} | jq -eMr '.channel') || { echo "docker channel not set"; exit 1; }

echo "REGISTRY ${REGISTRY}"
echo "CHANNEL ${CHANNEL}"
