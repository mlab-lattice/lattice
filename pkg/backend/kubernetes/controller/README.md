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


## Controller types

There are 3 different types of controllers:

- Lattice
  - lattice controllers watch lattice custom resources and generate more lattice custom resources
  - an example is the lattice-system-controller. This controller watches lattice Systems, and creates Services and other lattice custom resources only
- Kubernetes
  - kubernetes controllers watch lattice custom resources and generate kubernetes resources
  - an example is the kubernetes-component-build-controller. This controller watches lattice ComponentBuilds and creates kubernetes Jobs to run them
- Cloud
  - cloud controllers watch lattice custom resources and either provison cloud infrastructure or create lattice custom resources whose controllers provision cloud infrastructure
  - each supported cloud provider should implement the required cloud controllers
  - an example is the aws-service-controller
    - This controller watches for lattice Services, and creates lattice custom resources based on the Service whose controllers will provision cloud resources.
    - For example, if one of the Service's components' ports had external_access.public = true, the aws-service-controller would create an AwsElb custom resource.
    - The aws-elb-controller would then see the AwsElb custom resource and provision an ELB.
