# Overview

This document will attempt to give an overview of a Lattice system's architecture.

This document assumes familiarity with Kubernetes. A high level overview is available in the [Kubernetes docs](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/architecture.md).

Note that there are some differences between how lattice internally configures things based on if it is running locally or in a cloud environment. These differences will be annotated throughout.

## Master node

The initial state of a Lattice system consists solely of a `master` node.

This node runs the usual Kubernetes master components:

- `apiserver`
- `controller-manager`
- `etcd`

In addition, it runs two Lattice master components:

- `manager-api`
- `lattice-controller-manager`

It also has some information seeded including [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/) roles and [Custom Resource Definitions](https://kubernetes.io/docs/concepts/api-extension/custom-resources/).

### Custom Resources

We seed the following Custom Resource Definitions:

- componentbuilds.lattice.mlab.com (`ComponentBuild`)
- configs.lattice.mlab.com (`Config`)
- servicebuilds.lattice.mlab.com (`ServiceBuild`)
- services.lattice.mlab.com (`Service`)
- systembuilds.lattice.mlab.com (`SystemBuild`)
- systemrollouts.lattice.mlab.com (`SystemRollout`)
- systems.lattice.mlab.com (`System`)
- systemteardowns.lattice.mlab.com (`SystemTeardown`)


### Components

#### manager-api

The `manager-api` exposes the interface for interacting with Lattice. The `manager-api` uses the Custom Resources mentioned above to utilize the Kubernetes `apiserver` (and therefore `etcd`) as the backing store for Lattice information.

The `manager-api` can be thought of as the translation layer between Lattice and Kubernetes.

#### lattice-controller-manager

Lattice uses the same cascading controller pattern as Kubernetes to translate from high-level resources to the more granular primitives that implement them (bear with me, this will become more clear in the example below).

There are three types of `lattice-controllers`:

- Lattice
  - These controllers control Lattice Custom Resources that translate into other Lattice Custom Resources.
- Kubernetes
  - These controllers control Lattice Custom Resources that translate into Kubernetes Resources.
- Cloud
  - These controllers control provisioning and deprovisioning cloud resources based off Lattice Custom Resources.

The `lattice-controller-manager` is in charge of keeping all of the controllers up and running

## Example

### Object creation

Say we run `lattice roll-out-system --version v1.0.0`. This will:

- `POST` to the `manager-api`.
- The `manager-api` will first retrieve the `v1.0.0` tag from the system's configured definition git repo.
- Then the `manager-api` `POST`s to the Kubernetes `apiserver` creating a new `SystemBuild` object, which contains the system's definition.
- The `manager-api` will also `POST` creating a new `SystemRollout` object, which contains the newly created `SystemBuild`'s ID.

### Building

A few things happen now.

#### lattice-system-lifecycle controller

There is a `lattice-controller` called `lattice-system-lifecycle` which is in charge of rolling out and tearing down systems. When the `SystemRollout` object is created, the `lattice-system-lifecycle` controller will be notified:

- The `system-lifecycle` controller checks and sees that there are no other `SystemRollout` or `SystemTeardown` objects that are currently in progress, so it marks the created `SystemRollout` as `accepted`, which means that this `SystemRollout` is slated to be the rollout that will be attempted. If another rollout request were to come in, the `lattice-system-lifecycle` controller would make it `failed` since there is already a `SystemRollout` being processed.
- The `system-lifecycle` controller then checks on the state of the `SystemBuild` that the `SystemRollout` references. It sees that it is not yet complete, so there is no more work to be done, it simply must wait until the build has finished.

#### lattice-system-build controller

While this has all been happening in the `lattice-system-lifecycle` controller, another controller, the `lattice-system-build` controller has also been active.

- The `lattice-system-build` controller noticed the `SystemBuild` was created. It then crawls the definition tree, and for each service in the definition, creates a `ServiceBuild` resource, which includes the definition of the service.

#### lattice-service-build controller

The `lattice-service-build` controller then notices each of the `ServiceBuild`s that were created. For each of the `ServiceBuild`s it will:

- Look at each component in the service's definition and:
  - Take the `sha256` hash of the component definition's `build` section.
  - Check to see if a successful or in-progress `ComponentBuild` that already exists that is tagged with the hash
  - If not, create a new `ComponentBuild`

#### kubernetes-component-build controller

The `kubernetes-component-build` controller notices each of the `ComponentBuild`s that were created. For each it will:

- Create a [Kubernetes Job](https://kubernetes.io/docs/api-reference/v1.8/#job-v1-batch) that will run the component-builder (more on that below)
- Add a [toleration](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/) and [nodeSelector](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) to the job so that it will only run on nodes that have been tainted and labeled with the key `node-role.kubernetes.io/lattice-build`[[0]](local-difference-0)

#### cloud-component-build controller

The `cloud-component-build` controller notices there are pending `ComponentBuilds`[[1]](local-difference-1):

- It will check to see if it has already provisioned a build node, and if so nothing happens
- If there is no build node provisioned, the `cloud-component-build` controller will provision a new node that is both tainted and labeled with `node-role.kubernetes.io/lattice-build`
- Once the build node is created, Kubernetes will schedule the build job onto the node

#### component-builder

The `component-builder` job will then:

- If the build has a `git_repository`:
  - clone the repo and check out the appropriate `tag`/`commit`
  - build a docker image using either `language` or `base_docker_image` as the base and `command` as the command to build
- If the build has a `docker_image` and `cache` is `true`:
  - pull the docker image
- If we built a docker image for a `git_repository` or a `docker_image` with `cache`, push the container image to the container registry

#### Trickle back up

Once the `component-builder` job is complete:

- The `kubernetes-component-build` controller will notice, and update the `ComponentBuild.Status` to reflect if the job succeeded or failed.
- The `lattice-service-build` controller will wait until either one of its `ComponentBuild`s has failed or all have succeeded, and then will update the `ServiceBuild.Status` to reflect that
- Similarly, the `lattice-system-build` controller will wait until either one of its `ServiceBuild`s has failed or all have succeeded, and then will update the `SystemBuild.Status` to reflect that

### Rolling out

#### lattice-system-lifecycle controller

Once the `SystemBuild.Status` is updated, the `lattice-system-lifecycle` controller will:

- If the `SystemBuild` failed, fail the `SystemRollout`
- If the `SystemBuild` succeeded:
  - move the `SystemRollout` from `accepted` to `in-progress`
  - create a `System`, which includes information about each service in the definition, including the definition for the service and the artifacts created by the component builder for that service (e.g. the docker images)

#### lattice-system controller

The `lattice-system` controller notices the new `System`:

- For each service in `System.Spec.Services`, the `lattice-system` controller will create a `Service`

#### kubernetes-service controller

The `kubernetes-service` controller notices the new `Service` and:

- Creates a [Kubernetes deployment](https://kubernetes.io/docs/api-reference/v1.8/#deployment-v1beta2-apps)
  - has a [toleration](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/) and [nodeSelector](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) so that it will only run on nodes that have been tainted and labeled with `node-role.kubernetes.io/lattice-service=<Service.Name>`[[0]](local-difference-0)
  - has an [init-container](https://kubernetes.io/docs/api-reference/v1.8/#container-v1-core) to prepare iptables for envoy (more on this later), a [container](https://kubernetes.io/docs/api-reference/v1.8/#container-v1-core) for each component, and a container for envoy (more on this later)
- If the deployment contains no public component ports:
  - Creates a [Kubernetes headless service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) targeting the deployment
- Otherwise:
  - Creates a [Kubernetes NodePort service](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport) targeting the deployment

#### cloud-service controller

The `cloud-service` controller notices the new `Services`[[1]](local-difference-1):

- The `cloud-component-build` controller will provision new nodes that are both tainted and labeled with `node-role.kubernetes.io/lattice-service=<Service.Name>`
  - e.g. on AWS, creates an [autoscaling group](https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html)
- If the `Service` contains components with public ports, it will also provision a load balancer targeting the nodes and the port exposed by the NodePort service
- Add a DNS entry for `<SERVICE_PATH_DOMAIN>.system.internal`[[2]](local-difference-2)
  - e.g for a service `/foo/bar/buzz` add `buzz.bar.foo.system.internal`
  - for a normal stateless service, this DNS entry will be an `A` record mapping to the first IP of a configured CIDR block (more on this below)

#### service node

When the service node gets provisioned, Kubernetes will assign one of the Pods for the Deployment to it:
- First, our `initContainer` to prepare envoy will be run, and will:
  - create an `envoy` config file
  - via `iptables`, redirect all traffic from a configured CIDR block to the `localhost` port that `envoy` will be listening on
    - the effect of this is that when a service does a name look up on another lattice service (e.g. to the `buzz.bar.foo.system.internal` mentioned above) the traffic is trapped to `envoy`, which will know how to forward it along (more on this below)

As part of seeding Lattice, along side RBAC and the master components, there was also [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) running the `envoy-xds-api` on nodes tagged with `node-role.kubernetes.io/lattice-service`[[4]](local-difference-3):
- `envoy-xds-api` watches the [Kubernetes Endpoints](https://kubernetes.io/docs/api-reference/v1.8/#endpoints-v1-core) collection, and serves up information to the `envoy` running locally about what other services exist, and how to send traffic to them

Finally, once the `initContainer` succeeds, each component's container is run, alongside `envoy`, which in turn starts asking the local `envoy-xds-api` about the topology of the service mesh and in turn handling requests to/from the user components

#### Trickle back up

- Once a Deployments is successfully running, the `lattice-service` controller updates the `Service.Status.State` from `rolling-out` to `rolled-out` 
- Once all `Services.Status.State` are `rolled-out` the `lattice-system` controller updates the `System.Status.State` from `rolling-out` to `rolled-out`
- Once the `System.Status.State` is `rolled-out`, the `lattice-lifecycle` controller updates the `SystemRollout.Status.Statue` to `succeeded`

## Do it all again

Now lets say we updated one service in the system and wanted to roll it out. We would update the system to point at the new git tag/commit, and run `lattice roll-out-system --version v1.0.0`.

The exact same sequence of events would happen, with a few small changes:
- For every service in the system besides the one that changed, the `lattice-service-build` controller would find a succeeded `ComponenetBuild` that was tagged with the desired hash for the services, and would not have to create new `ComponentBuilds`
- The `lattice-lifecycle` controller would simply update the `System.Spec` instead of creating a new one
- The `lattice-system` controller would simply update the `Service.Spec` of the changed service
- The `lattice-service` controller would simply update the `Deployment.Spec`, and Kubernetes' `deployment` controller would take care of rolling out the new containers

A similar process would apply for something such as scaling up the number of instances of a service by changing its `resources.num_instances`:
- The `lattice-service` controller would update the `Deployment.Spec` to indicate more pods should be run
- The `cloud-service` controller would change the number of nodes for that service, and eventually the new node would spin up and the new Pod would be scheduled on it

# lattice-local differences

[local-difference-0]: Jobs and Deployments running locally are not given `nodeSelector`s since there is only one Kubernetes node in the local case

[local-difference-1]: Cloud controllers are not run in the local case. Everything is run on the single node.

[local-difference-2]: Eventually, the goal is to run a basic DNS server in the local case which watches the Lattice `Service` resources and the Kubernetes `Endpoint` resources, but we just add an entry for every single service into the `Pod.Spec.hostAlias` array

[local-difference-3]: As there is only one node in the local case, the `envoy-xds-api` DaemonSet is not applied with a `nodeSelector`
