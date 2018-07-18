#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
# no xtrace so we don't print the private key on each run

mkdir /root/.ssh
echo "$PRIVATE_KEY" > /root/.ssh/id_rsa
ssh-keyscan github.com > /root/.ssh/known_hosts
chmod 400 /root/.ssh/id_rsa

BINARY_DIRECTORY='../cli-binaries'
METADATA_DIRECTORY='../cli-metadata'
DOCGEN_BINARY='../docgen-binary/docgen'

if [ "$USE_TAG_FOR_VERSION" = "true" ]; then
    # get the version from git describe
    TAG_NAME=$(git describe)
else
    # get the version from
    TAG_NAME=$(cat ./.git/ref)
fi

LINUX_FILENAME=lattice_linux_amd64_v"$TAG_NAME"
DARWIN_FILENAME=lattice_darwin_amd64_v"$TAG_NAME"
LINUX_FILE="$BINARY_DIRECTORY"/"$LINUX_FILENAME"
DARWIN_FILE="$BINARY_DIRECTORY"/"$DARWIN_FILENAME"

# compile docgen binary
bazel --output_user_root=../cli-build-cache build --cpu k8 --features=static --features=pure //cmd/generate-latticectl-docs:generate-latticectl-docs
cp bazel-bin/cmd/generate-latticectl-docs/linux_amd64_static_pure_stripped/generate-latticectl-docs "$DOCGEN_BINARY"

# compile for linux
bazel --output_user_root=../cli-build-cache build --cpu k8 //cmd/latticectl --workspace_status_command=./scripts/workspace-status.sh
cp bazel-bin/cmd/cli/linux_amd64_stripped/latticectl "$LINUX_FILE"

# compile for macOS
bazel --output_user_root=../cli-build-cache build --experimental_platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 //cmd/latticectl --workspace_status_command=./scripts/workspace-status.sh
cp bazel-bin/cmd/cli/darwin_amd64_pure_stripped/latticectl "$DARWIN_FILE"

echo "$TAG_NAME" > "$METADATA_DIRECTORY"/tag
git tag -l -n "$TAG_NAME" | awk '{$1=""}1' | awk '{$1=$1}1' > "$METADATA_DIRECTORY"/tag_message
git show -s --format=%an "$TAG_NAME"^{commit} > "$METADATA_DIRECTORY"/tag_author
echo "$LINUX_FILENAME" > "$METADATA_DIRECTORY"/linux_filename
echo "$DARWIN_FILENAME" > "$METADATA_DIRECTORY"/darwin_filename
shasum -a 256 "$LINUX_FILE" | awk '{printf $1}' > "$METADATA_DIRECTORY"/linux_shasum
shasum -a 256 "$DARWIN_FILE" | awk '{printf $1}' > "$METADATA_DIRECTORY"/darwin_shasum
