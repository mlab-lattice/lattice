# Kubernetes

These documents give an overview of how lattice uses and extends kubernetes implement the `backend.Interface`. This documentation is mostly concerned with the provided facilities to extend Kubernetes, and how lattice leverages them.

It is assumed that you have read through the [architecture README](../README.md) and as such you are familiar with the concept of the `backend.Interface`.

We make an effort to explain the provided extension mechanics that we use, but may point out to external documentation. We do however assume a working knowledge of how vanilla Kubernetes works.

In particular, you should be familiar with:
- the kube api (links available in the [kube-api](kube-api.md) documentation, although practical experience is better)
- the controller paradigm used by Kubernetes
  - in particular, it is recommended you understand e.g. how a controller creates ReplicaSets for a Deployment, and how a different controller turns the ReplicaSet to a set of Pods

The recommended reading order is:

1. [kube-api](kube-api.md)
2. [custom-resources](custom-resources.md)
3. [controllers](controllers.md)
