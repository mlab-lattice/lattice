#!/usr/bin/env sh

set -e

CREDS_DIR=${WORK_DIR}/creds
if [ -d "${CREDS_DIR}" ]; then
    DOCKER_USER=$(cat ${CREDS_DIR}/docker | head -1)
    DOCKER_PASSWORD=$(cat ${CREDS_DIR}/docker | tail -1)

    docker login -u ${DOCKER_USER} -p ${DOCKER_PASSWORD} ${DOCKER_REGISTRY}
fi


cat > ${WORK_DIR}/Dockerfile <<EOF
FROM ${DOCKER_BASE_IMAGE}

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY src /usr/src/app

RUN ${BUILD_CMD}
EOF

DOCKER_FQN=${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:${DOCKER_IMAGE_TAG}

docker build ${WORK_DIR} -t ${DOCKER_FQN}

if [[ ${DOCKER_PUSH} -eq "1" ]]; then
    docker push ${DOCKER_FQN}
fi
