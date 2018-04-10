# Building Docker images

Prior to reading this, please read [building](building.md).

## Overview

Docker images are pushed to the `gcr.io/lattice-dev` repository.

There are multiple "channels" of docker images. A "channel" being a set of all necessary docker images built differently.

For example, there is the `stable` channel, which is images built from `master`. So the `envoy-prepare` image in the `stable` channel would be at `gcr.io/lattice-dev/stable-envoy-prepare`.

There is also the `stable-debug` channel (e.g. `gcr.io/lattice-dev/stable-debug-envoy-prepare`) which is the same images built with utilities such as `busybox` installed.

Each user also has their own channel. Currently the name of a user's channel is determined by the user-id on the build host.

For example:

```
$ whoami
kevinrosendahl
```

This would make my channel `gcr.io/lattice-dev/kevinrosendahl`. Each user also has a debug channel (e.g. `gcr.io/lattice-dev/kevinrosendahl-debug`).

The user channels exist so that different developers can develop using their own container images without effecting other developers' images.

The admin cli allows you to specify the `lattice-container-repo-prefix` for just this reason. For example to use the `kevinrosendahl-debug` channel, you would provision the system with:

```
$ bazel run -- //cmd/cli/admin provision local demo https://github.com/mlab-lattice/system__petflix \
      --backend-var lattice-container-registry=gcr.io/lattice-dev \
      --backend-var lattice-container-repo-prefix=kevinrosendahl-debug-
```

If using `api-services`, you can change the channel with:

```
LATTICE_CONTAINER_REPO_PREFIX="kevinrosendahl-debug-" npm start
```

## Building

### Auth

In order to be able to push to the `gcr.io/lattice-dev` repos, you must first be authorized and authenticated.

First ask Kevin to give you permissions the the `lattice-dev` repository.

Next, you'll need to download the [gcloud sdk](https://cloud.google.com/sdk/gcloud/) and log in by running

```
gcloud auth login
```

You'll then need to install the [docker-credential-helper](https://cloud.google.com/container-registry/docs/advanced-authentication#docker_credential_helper) by running:

```
gcloud components install docker-credential-gcr
```

If this doesn't work for some reason, you can manually install it by following the instructions at https://github.com/GoogleCloudPlatform/docker-credential-gcr

Then run:

```
docker-credential-gcr configure-docker
```

### Pushing

To build and push all images to your user channel, run

```
make docker.push-all-user
```

To build and push only a specific image to your user channel, run

```
make docker.push-image-user IMAGE=<IMAGE_NAME>
```

For example:

```
make docker.push-image-user IMAGE=envoy-prepare
```

In general, you should _not_ be pushing to the stable channel, but if you have to, run the same commands replacing "user" with "stable". For example:

```
make docker.push-all-images-stable
```
