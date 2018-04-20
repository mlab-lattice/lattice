# Building Docker images

Prior to reading this, please read [building](building.md).

## Dependencies

In addition to [Bazel](building.md), `jq` and `xz` are required to build docker images. These can be installed via brew for macOS (`brew install jq`, `brew install xz`).

## Overview

Lattice is composed of many different binaries (`api-server`, `lattice-controller-manager`, `component-builder`, etc). The different docker images that can be built can be found listed in the [Makefile](../../Makefile) above the `docker` targets.

## Saving and running

Often it's useful to be able to run the dockerized version of the binary locally. One such use case would be for the `lattice-controller-manager`, whose docker image comes bundled with terraform and the lattice terraform modules.

To save the image locally, run 

```shell
$ make docker.save IMAGE=<image>
```

or

```shell
$ make docker.save.<image>
```

The `lattice-controller-manager`'s image name is `kubernetes-lattice-controller-manager`.  To save it locally, you would run:

```shell
$ make docker.save IMAGE=kubernetes-lattice-controller-manager
```

or

```shell
$ make docker.save.kubernetes-lattice-controller-manager
```

This will build the docker image and tag it on the local docker daemon as `bazel/docker:<image>` (e.g. `bazel/docker:kubernetes-lattice-controller-manager`).

Often times you just want to be dropped into a shell in the docker image. For that you can run

```shell
$ make docker.run IMAGE=<image>
```

or

```shell
$ make docker.run.<image>
```

For example:

```shell
$ make docker.run.kubernetes-lattice-controller-manager
```

## Pushing

When pushing images, the docker images are labeled in the following format: `<REGISTRY>/<REPOSITORY-PREFIX>/<CHANNEL>/<IMAGE>`.

So for example, when pushing `latticectl` to the `gcr.io` registry, with the repository-prefix `lattice-dev` on the `testing` channel, the fully qualified name is `gcr.io/lattice-dev/testing/latticectl`.

To build and push a docker image, run

```shell
$ make docker.push IMAGE=<image> REGISTRY=<registry> REPOSITORY_PREFIX=<repository-prefix> CHANNEL=<channel>
```

or

```shell
$ make docker.push.<image> REGISTRY=<registry> REPOSITORY_PREFIX=<repository-prefix> CHANNEL=<channel>
```

For example, you could push to `gcr.io/lattice-dev/testing/latticectl` (assuming you have the proper credentials, see below) by running:

```shell
$ make docker.push.latticectl REGISTRY=gcr.io REPOSITORY_PREFIX=lattice-dev CHANNEL=testing
```

If you do not supply any of the arguments (`REGISTRY`, `REPOSITORY_PREFIX`, `CHANNEL`), it will look for a file `.lattice/docker.json` and look for the missing components in there.

For example, you could push to `gcr.io/lattice-dev/testing/latticectl` like so:

`.lattice/docker.json`:

```json
{
  "registry": "gcr.io",
  "repository_prefix": "lattice-dev"
}
```

```shell
$ make docker.push.latticectl CHANNEL=testing
```

This is the recommended way to configure the push target.


If you want to push all of the docker images instead of a single one, you can run:

```shell
$ make docker.push.all
```

In general, you should be pushing images to your own development channel. For example:

```shell
$ make docker.push.all CHANNEL=$(whoami)
```

### Stripped images

The base images used to build the lattice container images are from the [Distroless](https://github.com/GoogleContainerTools/distroless) project.

By default, the container images built use the [debug base image](https://github.com/GoogleContainerTools/distroless#debug-images), which is basically just the [base image](https://github.com/GoogleContainerTools/distroless/blob/master/base/README.md) plus busybox.

To use the non-debug base image for the containers, use the `-stripped` suffix for the `docker.push` targets.

For example:

```shell
$ make docker.push-stripped.latticectl
```

or

```shell
$ make docker.push-stripped.all
```

The stripped images are labeled in the following format: `<REGISTRY>/<REPOSITORY-PREFIX>/<CHANNEL>/<IMAGE>-stripped`, for example `gcr.io/lattice-dev/testing/latticectl-stripped`.

### Auth

In order to be able to push to a remote registry, you will need to be authenticated and authorized.

Here is how to get set up to push to Google Container Registry repository.

First, ask the project administrator to make your user a `Storage Admin` (https://cloud.google.com/container-registry/docs/access-control).

Next, you'll need to download the [gcloud sdk](https://cloud.google.com/sdk/gcloud/) and log in by running

```shell
$ gcloud auth login
```

You'll then need to install the [docker-credential-helper](https://cloud.google.com/container-registry/docs/advanced-authentication#docker_credential_helper) by running:

```shell
$ gcloud components install docker-credential-gcr
```

If this doesn't work for some reason, you can manually install it by following the instructions at https://github.com/GoogleCloudPlatform/docker-credential-gcr

Then run:

```shell
$ docker-credential-gcr configure-docker
```

You should now be able to push docker images to your `gcr.io` repository.
