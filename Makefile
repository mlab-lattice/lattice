# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CLOUD_IMAGE_DIR = $(DIR)/cloud-images
CLOUD_IMAGE_BUILD_DIR = $(CLOUD_IMAGE_DIR)/build
CLOUD_IMAGE_BUILD_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/build
CLOUD_IMAGE_AWS_SYSTEM_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/aws/$(LATTICE_SYSTEM_ID)

LOCAL_REGISTRY = lattice-local
DEV_REGISTRY = gcr.io/lattice-dev
DEV_TAG ?= latest

CONTAINER_NAME_BUILD = lattice-system-builder

# Basic build/clean/test
.PHONY: build
build: gazelle
	@bazel build //...:all

.PHONY: clean
clean:
	@bazel clean

.PHONY: test
test: gazelle
	@bazel test --test_output=errors //...

.PHONY: gazelle
gazelle:
	@bazel run //:gazelle

# docker
.PHONY: docker-push-image
docker-push-image:
	bazel run //docker:push-$(IMAGE)
	bazel run //docker:push-debug-$(IMAGE)

# local binaries
.PHONY: update-binaries
update-binaries: update-binary-cli-admin update-binary-cli-user

.PHONY: update-binary-cli-admin
update-binary-cli-admin:
	@bazel build //cmd/cli/admin
	cp -f $(DIR)/bazel-bin/cmd/cli/admin/admin $(DIR)/bin/lattice-admin

.PHONY: update-binary-cli-user
update-binary-cli-user:
	@bazel build //cmd/cli/user
	cp -f $(DIR)/bazel-bin/cmd/cli/user/user $(DIR)/bin/lattice-system

# docker build hackery
.PHONY: docker-hack-enter-build-shell
docker-hack-enter-build-shell: docker-hack-build-start-build-container
	docker exec -it $(CONTAINER_NAME_BUILD) ./docker/bazel-builder/wrap-creds-and-exec.sh /bin/bash

.PHONY: docker-hack-push-image
docker-hack-push-image: docker-hack-build-start-build-container
	docker exec $(CONTAINER_NAME_BUILD) ./docker/bazel-builder/wrap-creds-and-exec.sh make docker-push-image IMAGE=$(IMAGE)

.PHONY: docker-hack-push-all-images
docker-hack-push-all-images:
	make docker-hack-push-image IMAGE=envoy-prepare
	make docker-hack-push-image IMAGE=kubernetes-bootstrap-lattice
	make docker-hack-push-image IMAGE=kubernetes-component-builder
	make docker-hack-push-image IMAGE=kubernetes-envoy-xds-api-rest-per-node
	make docker-hack-push-image IMAGE=kubernetes-lattice-controller-manager
	make docker-hack-push-image IMAGE=kubernetes-manager-api-rest

.PHONY: docker-hack-build-bazel-build
docker-hack-build-bazel-build:
	docker build $(DIR)/docker -f $(DIR)/docker/bazel-builder/Dockerfile.bazel-build -t lattice-build/bazel-build

.PHONY: docker-hack-build-start-build-container
docker-hack-build-start-build-container: docker-hack-build-bazel-build
	$(DIR)/docker/bazel-builder/start-build-container.sh

# cloud images
.PHONY: cloud-images-build
cloud-images-build: cloud-images-build-base-node-image cloud-images-build-master-node-image

.PHONY: cloud-images-build-base-node-image
cloud-images-build-base-node-image:
	$(CLOUD_IMAGE_BUILD_DIR)/build-base-node-image

.PHONY: cloud-images-build-master-node-image
cloud-images-build-master-node-image:
	$(CLOUD_IMAGE_BUILD_DIR)/build-master-node-image

.PHONY: cloud-images-clean
cloud-images-clean:
	rm -rf $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts

.PHONY: cloud-images-clean-master-node-image
cloud-images-clean-master-node-image:
	rm -rf $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts/master-node
	rm -f $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts/master-node-ami-id
