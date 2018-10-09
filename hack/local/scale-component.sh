#!/usr/bin/env bash

set -u
set -e

echo "halting ${@}"
kubectl scale -n lattice-internal deployment.apps/${@} --replicas 0

echo "waiting for ${@} to be halt"
while [[ $(kubectl get -n lattice-internal deployment.apps/${@} -o json | jq ".status.availableReplicas") != "null" ]]; do
    sleep 1
done
