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

BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH = debian-with-ssh
BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH_DEV = $(DEV_REGISTRY)/$(BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH):$(DEV_TAG)

BASE_DOCKER_IMAGE_UBUNTU_WITH_IPTABLES = ubuntu-with-iptables
BASE_DOCKER_IMAGE_DEBIAN_WITH_IPTABLES_DEV = $(DEV_REGISTRY)/$(BASE_DOCKER_IMAGE_UBUNTU_WITH_IPTABLES):$(DEV_TAG)

BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS = ubuntu-with-aws
BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS_DEV = $(DEV_REGISTRY)/$(BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS):$(DEV_TAG)

DOCKER_IMAGE_COMPONENT_BUILD_BUILD = component-build-build-docker-image
DOCKER_IMAGE_COMPONENT_BUILD_BUILD_BAZEL = bazel/docker:$(DOCKER_IMAGE_COMPONENT_BUILD_BUILD)
DOCKER_IMAGE_COMPONENT_BUILD_BUILD_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_COMPONENT_BUILD_BUILD):$(DEV_TAG)
DOCKER_IMAGE_COMPONENT_BUILD_BUILD_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_COMPONENT_BUILD_BUILD)

DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS = component-build-get-ecr-creds
DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS_BAZEL = bazel/docker:$(DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS)
DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS):$(DEV_TAG)
DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS)

DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO = component-build-pull-git-repo
DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_BAZEL = bazel/docker:$(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO)
DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO):$(DEV_TAG)
DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO)

DOCKER_IMAGE_ENVOY_PREPARE_ENVOY = envoy-prepare-envoy
DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_BAZEL = bazel/docker:$(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY)
DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY):$(DEV_TAG)
DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY)

DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE = kubernetes-bootstrap-lattice
DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_BAZEL = bazel/docker:$(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE)
DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE):$(DEV_TAG)
DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE)

DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE = kubernetes-envoy-xds-api-rest-per-node
DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_BAZEL = bazel/docker:$(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE)
DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE):$(DEV_TAG)
DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE)

DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER = kubernetes-lattice-controller-manager
DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_BAZEL = bazel/docker:$(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER)
DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER):$(DEV_TAG)
DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER)

DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST = kubernetes-manager-api-rest
DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_BAZEL = bazel/docker:$(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST)
DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST):$(DEV_TAG)
DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST)

DOCKER_IMAGE_LATTICE_SYSTEM_CLI = lattice-system-cli
DOCKER_IMAGE_LATTICE_SYSTEM_CLI_BAZEL = bazel/docker:$(DOCKER_IMAGE_LATTICE_SYSTEM_CLI)
DOCKER_IMAGE_LATTICE_SYSTEM_CLI_DEV = $(DEV_REGISTRY)/$(DOCKER_IMAGE_LATTICE_SYSTEM_CLI):$(DEV_TAG)
DOCKER_IMAGE_LATTICE_SYSTEM_CLI_LOCAL = $(LOCAL_REGISTRY)/$(DOCKER_IMAGE_LATTICE_SYSTEM_CLI)


# Basic build/clean/test
.PHONY: build
build: gazelle
	@bazel build //...:all

.PHONY: build-docker-images
build-docker-images: build-docker-images-sh build-docker-images-go

.PHONY: clean
clean:
	@bazel clean

.PHONY: test
test: gazelle
	@bazel test --test_output=errors //...

.PHONY: gazelle
gazelle:
	@bazel run //:gazelle

.PHONY: docker-build-base-images
docker-build-base-images:
	docker build $(DIR)/docker/component-build -f $(DIR)/docker/component-build/Dockerfile.aws -t $(BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS_DEV)
	docker build $(DIR)/docker/debian -f $(DIR)/docker/debian/Dockerfile.iptables -t $(BASE_DOCKER_IMAGE_DEBIAN_WITH_IPTABLES_DEV)
	docker build $(DIR)/docker/debian -f $(DIR)/docker/debian/Dockerfile.ssh -t $(BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH_DEV)

.PHONY: docker-push-dev-base-images
docker-push-dev-base-images:
	gcloud docker -- push $(BASE_DOCKER_IMAGE_DEBIAN_WITH_IPTABLES_DEV)
	gcloud docker -- push $(BASE_DOCKER_IMAGE_DEBIAN_WITH_SSH_DEV)
	gcloud docker -- push $(BASE_DOCKER_IMAGE_UBUNTU_WITH_AWS_DEV)

.PHONY: docker-build-and-push-dev-base-images
docker-build-and-push-dev-base-images: docker-build-base-images docker-push-dev-base-images

.PHONY: build-docker-images-sh
build-docker-images-sh:
	@bazel run //docker:component-build-build-docker-image
	@bazel run //docker:component-build-get-ecr-creds
	@bazel run //docker:component-build-pull-git-repo
	@bazel run //docker:envoy-prepare-envoy

.PHONY: build-docker-images-go
build-docker-images-go: build-docker-image-kubernetes-bootstrap-lattice \
						build-docker-image-kubernetes-envoy-xds-api-rest-per-node \
						build-docker-image-kubernetes-lattice-controller-manager \
						build-docker-image-kubernetes-manager-api-rest \
						build-docker-image-lattice-system-cli

.PHONY: build-docker-image-kubernetes-bootstrap-lattice
build-docker-image-kubernetes-bootstrap-lattice: gazelle
	@bazel run //docker:kubernetes-bootstrap-lattice -- --norun

.PHONY: build-docker-image-kubernetes-envoy-xds-api-rest-per-node
build-docker-image-kubernetes-envoy-xds-api-rest-per-node: gazelle
	@bazel run //docker:kubernetes-envoy-xds-api-rest-per-node -- --norun

.PHONY: build-docker-image-kubernetes-lattice-controller-manager
build-docker-image-kubernetes-lattice-controller-manager: gazelle
	@bazel run //docker:kubernetes-lattice-controller-manager -- --norun

.PHONY: build-docker-image-kubernetes-manager-api-rest
build-docker-image-kubernetes-manager-api-rest: gazelle
	@bazel run //docker:kubernetes-manager-api-rest -- --norun

.PHONY: build-docker-image-lattice-system-cli
build-docker-image-lattice-system-cli: gazelle
	@bazel run //docker:lattice-system-cli -- --norun


# local binaries
.PHONY: update-local-binary-cli
update-local-binary-cli:
	@bazel build //cmd/cli
	cp -f $(DIR)/bazel-bin/cmd/cli/cli $(DIR)/bin/lattice-system

# docker build hackery
.PHONY: docker-enter-build-shell
docker-enter-build-shell: docker-build-start-build-container
	docker exec -it $(CONTAINER_NAME_BUILD) ./docker/bazel-builder/wrap-creds-and-exec.sh /bin/bash

.PHONY: docker-build-bazel-build
docker-build-bazel-build:
	docker build $(DIR)/docker -f $(DIR)/docker/bazel-builder/Dockerfile.bazel-build -t lattice-build/bazel-build

.PHONY: docker-build-start-build-container
docker-build-start-build-container: docker-build-bazel-build
	$(DIR)/docker/bazel-builder/start-build-container.sh

# docker save
.PHONY: docker-build-and-save
docker-build-and-save: docker-build docker-save

.PHONY: docker-save
docker-save: docker-save-component-build-build-docker-image \
			 docker-save-component-build-pull-git-repo \
			 docker-save-envoy-prepare-envoy \
			 docker-save-kubernetes-bootstrap-lattice \
			 docker-save-kubernetes-envoy-xds-api-rest-per-node \
			 docker-save-kubernetes-lattice-controller-manager \
			 docker-save-lattice-system-cli

.PHONY: docker-save-component-build-build-docker-image
docker-save-component-build-build-docker-image: docker-tag-local-component-build-build-docker-image
	docker save $(DOCKER_IMAGE_COMPONENT_BUILD_BUILD_LOCAL) -o $(dest)/$(DOCKER_IMAGE_COMPONENT_BUILD_BUILD)

.PHONY: docker-save-component-build-pull-git-repo
docker-save-component-build-pull-git-repo: docker-tag-local-component-build-pull-git-repo
	docker save $(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_LOCAL) -o $(dest)/$(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO)

.PHONY: docker-save-envoy-prepare-envoy
docker-save-envoy-prepare-envoy: docker-tag-local-envoy-prepare-envoy
	docker save $(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_LOCAL) -o $(dest)/$(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY)

.PHONY: docker-save-kubernetes-bootstrap-lattice
docker-save-kubernetes-bootstrap-lattice: docker-tag-local-kubernetes-bootstrap-lattice
	docker save $(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_LOCAL) -o $(dest)/$(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE)

.PHONY: docker-save-kubernetes-envoy-xds-api-rest-per-node
docker-save-kubernetes-envoy-xds-api-rest-per-node: docker-tag-local-kubernetes-envoy-xds-api-rest-per-node
	docker save $(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_LOCAL) -o $(dest)/$(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE)

.PHONY: docker-save-kubernetes-lattice-controller-manager
docker-save-kubernetes-lattice-controller-manager: docker-tag-local-kubernetes-lattice-controller-manager
	docker save $(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_LOCAL) -o $(dest)/$(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER)

.PHONY: docker-save-kubernetes-manager-api-rest
docker-save-kubernetes-manager-api-rest: docker-tag-local-kubernetes-lattice-controller-manager
	docker save $(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_LOCAL) -o $(dest)/$(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST)

.PHONY: docker-save-lattice-system-cli
docker-save-lattice-system-cli: docker-tag-local-lattice-system-cli
	docker save $(DOCKER_IMAGE_LATTICE_SYSTEM_CLI_LOCAL) -o $(dest)/$(DOCKER_IMAGE_LATTICE_SYSTEM_CLI)

.PHONY: docker-tag-local-component-build-build-docker-image
docker-tag-local-component-build-build-docker-image:
	docker tag $(DOCKER_IMAGE_COMPONENT_BUILD_BUILD_BAZEL) $(DOCKER_IMAGE_COMPONENT_BUILD_BUILD_LOCAL)

.PHONY: docker-tag-local-component-build-pull-git-repo
docker-tag-local-component-build-pull-git-repo:
	docker tag $(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_BAZEL) $(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_LOCAL)

.PHONY: docker-tag-local-envoy-prepare-envoy
docker-tag-local-envoy-prepare-envoy:
	docker tag $(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_BAZEL) $(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_LOCAL)

.PHONY: docker-tag-local-kubernetes-bootstrap-lattice
docker-tag-local-kubernetes-bootstrap-lattice:
	docker tag $(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_BAZEL) $(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_LOCAL)

.PHONY: docker-tag-local-kubernetes-envoy-xds-api-rest-per-node
docker-tag-local-kubernetes-envoy-xds-api-rest-per-node:
	docker tag $(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_BAZEL) $(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_LOCAL)

.PHONY: docker-tag-local-kubernetes-lattice-controller-manager
docker-tag-local-kubernetes-lattice-controller-manager:
	docker tag $(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_BAZEL) $(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_LOCAL)

.PHONY: docker-tag-local-kubernetes-manager-api-rest
docker-tag-local-kubernetes-manager-api-rest:
	docker tag $(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_BAZEL) $(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_LOCAL)

.PHONY: docker-tag-local-lattice-system-cli
docker-tag-local-lattice-system-cli:
	docker tag $(DOCKER_IMAGE_LATTICE_SYSTEM_CLI_BAZEL) $(DOCKER_IMAGE_LATTICE_SYSTEM_CLI_LOCAL)


# docker push-dev
.PHONY: docker-push-dev
docker-push-dev: docker-push-dev-component-build-build-docker-image \
				 docker-push-dev-component-build-get-ecr-creds \
				 docker-push-dev-component-build-pull-git-repo \
				 docker-push-dev-envoy-prepare-envoy \
				 docker-push-dev-kubernetes-bootstrap-lattice \
 				 docker-push-dev-kubernetes-envoy-xds-api-rest-per-node \
				 docker-push-dev-kubernetes-lattice-controller-manager \
				 docker-push-dev-kubernetes-manager-api-rest \
				 docker-push-dev-lattice-system-cli

.PHONY: docker-build-and-push-dev
docker-build-and-push-dev: docker-build docker-push-dev

.PHONY: docker-build-and-push-dev-kubernetes-master-components
docker-build-and-push-dev-kubernetes-master-components: docker-build-kubernetes-master-components \
														docker-push-dev-kubernetes-master-components

.PHONY: docker-push-dev-kubernetes-master-components
docker-push-dev-kubernetes-master-components: docker-tag-dev-kubernetes-master-components \
											  docker-push-dev-kubernetes-lattice-controller-manager \
											  docker-push-dev-kubernetes-manager-api-rest

.PHONY: docker-push-dev-component-build-build-docker-image
docker-push-dev-component-build-build-docker-image: docker-tag-dev-component-build-build-docker-image
	gcloud docker -- push $(DOCKER_IMAGE_COMPONENT_BUILD_BUILD_DEV)

.PHONY: docker-push-dev-component-build-get-ecr-creds
docker-push-dev-component-build-get-ecr-creds: docker-tag-dev-component-build-get-ecr-creds
	gcloud docker -- push $(DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS_DEV)

.PHONY: docker-push-dev-component-build-pull-git-repo
docker-push-dev-component-build-pull-git-repo: docker-tag-dev-component-build-pull-git-repo
	gcloud docker -- push $(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_DEV)

.PHONY: docker-push-dev-envoy-prepare-envoy
docker-push-dev-envoy-prepare-envoy: docker-tag-dev-envoy-prepare-envoy
	gcloud docker -- push $(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_DEV)

.PHONY: docker-push-dev-kubernetes-bootstrap-lattice
docker-push-dev-kubernetes-bootstrap-lattice: docker-tag-dev-kubernetes-bootstrap-lattice
	gcloud docker -- push $(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_DEV)

.PHONY: docker-push-dev-kubernetes-envoy-xds-api-rest-per-node
docker-push-dev-kubernetes-envoy-xds-api-rest-per-node: docker-tag-dev-kubernetes-envoy-xds-api-rest-per-node
	gcloud docker -- push $(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_DEV)

.PHONY: docker-push-dev-kubernetes-lattice-controller-manager
docker-push-dev-kubernetes-lattice-controller-manager: docker-tag-dev-kubernetes-lattice-controller-manager
	gcloud docker -- push $(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_DEV)

.PHONY: docker-push-dev-kubernetes-manager-api-rest
docker-push-dev-kubernetes-manager-api-rest: docker-tag-dev-kubernetes-manager-api-rest
	gcloud docker -- push $(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_DEV)

.PHONY: docker-push-dev-lattice-system-cli
docker-push-dev-lattice-system-cli: docker-tag-dev-lattice-system-cli
	gcloud docker -- push $(DOCKER_IMAGE_LATTICE_SYSTEM_CLI_DEV)

.PHONY: docker-tag-dev-kubernetes-master-components
docker-tag-dev-kubernetes-master-components: docker-tag-dev-kubernetes-lattice-controller-manager \
											 docker-tag-dev-kubernetes-manager-api-rest

.PHONY: docker-tag-dev-component-build-build-docker-image
docker-tag-dev-component-build-build-docker-image:
	docker tag $(DOCKER_IMAGE_COMPONENT_BUILD_BUILD_BAZEL) $(DOCKER_IMAGE_COMPONENT_BUILD_BUILD_DEV)

.PHONY: docker-tag-dev-component-build-get-ecr-creds
docker-tag-dev-component-build-get-ecr-creds:
	docker tag $(DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS_BAZEL) $(DOCKER_IMAGE_COMPONENT_BUILD_GET_ECR_CREDS_DEV)

.PHONY: docker-tag-dev-component-build-pull-git-repo
docker-tag-dev-component-build-pull-git-repo:
	docker tag $(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_BAZEL) $(DOCKER_IMAGE_COMPONENT_BUILD_PULL_GIT_REPO_DEV)

.PHONY: docker-tag-dev-envoy-prepare-envoy
docker-tag-dev-envoy-prepare-envoy:
	docker tag $(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_BAZEL) $(DOCKER_IMAGE_ENVOY_PREPARE_ENVOY_DEV)

.PHONY: docker-tag-dev-kubernetes-bootstrap-lattice
docker-tag-dev-kubernetes-bootstrap-lattice:
	docker tag $(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_BAZEL) $(DOCKER_IMAGE_KUBERNETES_BOOTSTRAP_LATTICE_DEV)

.PHONY: docker-tag-dev-kubernetes-envoy-xds-api-rest-per-node
docker-tag-dev-kubernetes-envoy-xds-api-rest-per-node:
	docker tag $(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_BAZEL) $(DOCKER_IMAGE_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_DEV)

.PHONY: docker-tag-dev-kubernetes-lattice-controller-manager
docker-tag-dev-kubernetes-lattice-controller-manager:
	docker tag $(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_BAZEL) $(DOCKER_IMAGE_KUBERNETES_LATTICE_CONTROLLER_MANAGER_DEV)

.PHONY: docker-tag-dev-kubernetes-manager-api-rest
docker-tag-dev-kubernetes-manager-api-rest:
	docker tag $(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_BAZEL) $(DOCKER_IMAGE_KUBERNETES_MANAGER_API_REST_DEV)

.PHONY: docker-tag-dev-lattice-system-cli
docker-tag-dev-lattice-system-cli:
	docker tag $(DOCKER_IMAGE_LATTICE_SYSTEM_CLI_BAZEL) $(DOCKER_IMAGE_LATTICE_SYSTEM_CLI_DEV)


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
