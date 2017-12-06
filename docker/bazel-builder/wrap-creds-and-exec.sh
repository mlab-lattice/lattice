#!/usr/bin/env bash

set -e

mkdir -p ~/.ssh
eval "$(ssh-agent -s)"
ssh-add /tmp/.ssh/id_rsa-github
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

cd /src
eval ${@}
