#!/usr/bin/env bash

set -u
set -e

echo "killing ${@}"
kubectl delete -n lattice-internal pod -l control-plane.lattice.mlab.com/service=${@}

# wait a bit so the deployment's availableReplicas will be updated
sleep 2

echo "waiting for ${@} to be available"
while [[ $(kubectl get -n lattice-internal deployment.apps/${@} -o json | jq ".status.availableReplicas") == "0" ]]; do
    sleep 1
done
