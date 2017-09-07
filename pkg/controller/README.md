# Controllers

This package contains the different controllers in charge of the lattice custom resources.

## Prerequisites
__IMPORTANT__

Please read the following documentation before looking at this code:

- https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/controller-ref.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/garbage-collection.md
- https://github.com/kubernetes/community/blob/master/contributors/design-proposals/controller_history.md

You should also familiarize yourself with [client-go](https://github.com/kubernetes/client-go). Particularly [informers](https://godoc.org/k8s.io/client-go/informers) and [cache](https://godoc.org/k8s.io/client-go/tools/cache).

You should read through and understand the following as well:
- https://github.com/kubernetes/kubernetes/tree/master/cmd/kube-controller-manager
- https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/replicaset
- https://github.com/kubernetes/kubernetes/tree/master/pkg/controller/deployment