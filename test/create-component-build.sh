#!/usr/bin/env bash

cat <<EOF | kubectl create -f -
apiVersion: lattice.mlab.com/v1
kind: ComponentBuild
metadata:
  name: ${1}
  namespace: default
spec:
  definitionBlock:
    command: npm install
    git_repository:
      commit: 16d0ad5a7ef969b34174c39f12a588a38f4ff076
      url: https://github.com/kevindrosendahl/example__hello-world-service-chaining
    language: node:boron
EOF
