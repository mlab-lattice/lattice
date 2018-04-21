#!/usr/bin/env bash

set -u

go get -d k8s.io/code-generator

set -e

cd ${GOPATH}/src/k8s.io/code-generator
git fetch origin
git checkout kubernetes-${KUBERNETES_VERSION}
./generate-groups.sh all \
                     github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated \
                     github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis \
                     lattice:v1 \
                     --go-header-file ~/go/src/github.com/mlab-lattice/lattice/hack/kubernetes/codegen/go-header.txt
