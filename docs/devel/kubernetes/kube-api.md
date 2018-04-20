# Kubernetes API

It's recommended to go read the following:
- https://kubernetes.io/docs/reference/api-overview/
- https://kubernetes.io/docs/reference/api-concepts/

The general overview is that the kubernetes API is itself made up of many different versioned APIs. Different resources live under these different APIs.

For example, the `Deployment` resource lives under the `apps` API, which has three versions, `v1beta1`, `v1beta2`, and `v1`.

In Kubernetes, you create a resource, and the controllers will work towards making that resource exist in the state described by the resource.

For example, when you create a `Deployment`, the deployment controller, which is managed by the kubernetes `controller-manager`, will see the new `Deployment`, and create a `ReplicaSet` that matches the `Deployment`'s spec. The replica set controller will then act upon the new `ReplicaSet` resource, etc.

Resources _generally_ have the following:
 
- apiVersion: which API and which version of the API the resource belongs to
- kind: the type within the API of the resource
- namespace: most resources have a namespace, but some like the `Node` resource in the `core` API are cluster-wide
- name: must be unique within the namespace
- labels: "key/value pairs that are attached to objects. labels are intended to be used to specify identifying attributes of objects that are meaningful and relevant to users, but do not directly imply semantics to the core system" is what is written in the kubernetes docs. in practice labels do effect the core system
- annotations: 
- spec: the desired state of the resource that the controllers will work towards
- status: the current status of the resource
 
These are the main important parts of a resource, but there are a few other that are noteworthy:

- uid: a value that is unique across time and space. for example, if you create a resource in a namespace then delete it and recreate it (i.e. make a new resource with the same name), it will have a new uid.
- deletionTimestamp: if not null, indicates that the resource has been deleted and is in the process of being cleaned up
- finalizers: a list that must be empty before the resource can be deleted. this is useful if a controller has to do some cleanup before a resource is deleted. when the controller first operates on the resource, it adds itself to the list of finalizers. then when the resource gets deleted (i.e. deletionTimestamp is not null), the controller does its cleanup. when it's finished it removes itself from the list of finalizers. when the list of finalizers is empty, the resource is fully deleted
- ownerReferences: a list of resources that "own" this resource. this can allow for automatic [garbage collection](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/)
- generation: an integer incremented by the api when the resource is changed. useful in combination with an `observedGeneration` field in the resource's status that is updated by the controller. You can compare `generation` with `observedGeneration` to see if the version of the resource in its spec has been seen and acted upon by its controller
 