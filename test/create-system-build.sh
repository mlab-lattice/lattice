#!/usr/bin/env bash

cat <<EOF | kubectl create -f -
apiVersion: lattice.mlab.com/v1
kind: SystemBuild
metadata:
  name: ${1}
  namespace: default
  labels:
    build.system.lattice.mlab.com/version: v1.0.0
spec:
  services:
  - path: "/a/b/c"
    definition:
      components:
      - name: http
        build:
          command: npm install
          git_repository:
            commit: 16d0ad5a7ef969b34174c39f12a588a38f4ff076
            url: https://github.com/kevindrosendahl/example__hello-world-service-chaining
          language: node:boron
EOF
