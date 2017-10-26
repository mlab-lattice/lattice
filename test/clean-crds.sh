#!/usr/bin/env bash

kubectl delete componentbuilds.lattice.mlab.com  --all --namespace=lattice-internal
#kubectl delete configs.lattice.mlab.com          --all --namespace=lattice-internal
kubectl delete servicebuilds.lattice.mlab.com    --all --namespace=lattice-internal
kubectl delete services.lattice.mlab.com         --all --namespace=lattice-user-system
kubectl delete systembuilds.lattice.mlab.com     --all --namespace=lattice-internal
kubectl delete systemrollouts.lattice.mlab.com   --all --namespace=lattice-internal
kubectl delete systems.lattice.mlab.com          --all --namespace=lattice-user-system
