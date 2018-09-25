#!/usr/bin/env bash


DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "seeding resources"
kubectl create namespace lattice-internal

set -e
helm template ${DIR} ${@} | kubectl apply -f -

echo "waiting for api to be available"
while ! [[ $(kubectl get -n lattice-internal deployment.apps/api-server -o json | jq ".status.availableReplicas") == "1" ]]; do
    sleep 1
done

echo "api is running"