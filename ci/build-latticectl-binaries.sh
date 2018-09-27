#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
# no xtrace so we don't print the private key on each run

mkdir /root/.ssh
echo ${PRIVATE_KEY} > /root/.ssh/id_rsa
ssh-keyscan github.com > /root/.ssh/known_hosts
chmod 400 /root/.ssh/id_rsa

BINARY_DIRECTORY='../cli-binaries'
METADATA_DIRECTORY='../cli-metadata'

if [ ${USE_TAG_FOR_VERSION} == true ]; then
    # get the version from git describe
    TAG_NAME=$(git describe)
else
    # get the version from git
    TAG_NAME=$(git rev-parse --short HEAD)
fi

make build.platform.all \
    OUTPUT_USER_ROOT=../cli-build-cache \
    TARGET=//cmd/latticectl

declare -a os_list=(
    "darwin"
    "linux"
)

declare -a arch_list=(
    "amd64"
)

for os in "${os_list[@]}"
do
    for arch in "${arch_list[@]}"
    do
        DEST=${BINARY_DIRECTORY}/latticectl_${os}_${arch}_v${TAG_NAME}
        cp bazel-bin/cmd/latticectl/${os}_${arch}_pure_stripped/latticectl ${DEST}
        echo ${DEST} > ${METADATA_DIRECTORY}/${DEST}_filename
        shasum -a 256 ${DEST} | awk '{printf $1}' > ${METADATA_DIRECTORY}/${os}_shasum
    done
done

echo ${TAG_NAME} > ${METADATA_DIRECTORY}/tag
git tag -l -n ${TAG_NAME} | awk '{$1=""}1' | awk '{$1=$1}1' > ${METADATA_DIRECTORY}/tag_message
git show -s --format=%an ${TAG_NAME}^{commit} > ${METADATA_DIRECTORY}/tag_author
