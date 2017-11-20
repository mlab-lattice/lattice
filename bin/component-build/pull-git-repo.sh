#!/usr/bin/env sh

set -e

if [[ ! -z ${GIT_SSH_KEY_PATH} ]]; then
    eval "$(ssh-agent -s)"
    ssh-add -K ${GIT_SSH_KEY_PATH}
fi

REPO_DIR=${WORK_DIR}/src

if [ ! -d "${REPO_DIR}" ]; then
  git clone --verbose ${GIT_URL} ${REPO_DIR}
fi

cd ${REPO_DIR}
git checkout ${GIT_CHECKOUT_TARGET}
