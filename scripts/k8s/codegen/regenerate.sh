#!/usr/bin/env bash

set -e
set -u

go get -d k8s.io/code-generator
cd ${GOPATH}/src/k8s.io/code-generator
git checkout ${KUBERNETES_VERSION}
./generate-groups.sh all \
                     github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated \
                     github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis \
                     lattice:v1 \
                     --go-header-file ~/go/src/github.com/mlab-lattice/system/scripts/k8s/codegen/go-header.txt
