#!/usr/bin/env bash

set -e

# If REGISTRY or CHANNEL are set, uses the set values.
# Otherwise, looks for a file .lattice/docker.json at the root of the repository
# and tries to parse the values from the registry and channel keys respectively

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LATTICE_CONFIG_ROOT=${DIR}/../../.lattice

DOCKER_CONFIG_FILE_NAME=docker.json
DOCKER_CONFIG_FILE_PATH=${LATTICE_CONFIG_ROOT}/${DOCKER_CONFIG_FILE_NAME}

function config_file_value {
    if [ ! -f ${DOCKER_CONFIG_FILE_PATH} ]; then
        echo "${DOCKER_CONFIG_FILE_NAME} does not exist"; exit 1
    fi

    CONFIG=$(cat ${DOCKER_CONFIG_FILE_PATH})

    command -v jq >/dev/null 2>&1 || { >&2 echo "jq not installed"; exit 1; }
    echo ${CONFIG} | jq -e . >/dev/null 2>&1 || { >&2 echo "${DOCKER_CONFIG_FILE_NAME} is not valid JSON"; exit 1; }
    VALUE=$(echo ${CONFIG} | jq -eMr ".${1}") || { >&2 echo "${1} not set"; exit 1; }
    echo ${VALUE}
}

if [[ -z ${REGISTRY} ]]; then
    REGISTRY=$(config_file_value registry)
fi

if [[ -z ${REPOSITORY_PREFIX} ]]; then
    REPOSITORY_PREFIX=$(config_file_value repository_prefix)
fi

if [[ -z ${CHANNEL} ]]; then
    CHANNEL=$(config_file_value channel)
fi

echo "REGISTRY ${REGISTRY}"
echo "REPOSITORY_PREFIX ${REPOSITORY_PREFIX}"
echo "CHANNEL ${CHANNEL}"
