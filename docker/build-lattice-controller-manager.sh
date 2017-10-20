#!/usr/bin/env bash

set -e

eval "$(ssh-agent -s)"
ssh-add /root/.ssh/id_rsa-github
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

cd /src
/root/bin/bazel run //cmd/controller-manager:go_image -- --norun
