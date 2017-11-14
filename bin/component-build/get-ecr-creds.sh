#!/usr/bin/env sh

set -e

CREDS_DIR=${WORK_DIR}/creds
mkdir -p ${CREDS_DIR}
aws --region ${REGION} ecr get-login | awk '{print $4 "\n" $6}' > ${CREDS_DIR}/docker
