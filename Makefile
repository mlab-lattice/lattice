# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
BOOTSTRAP_DIR = $(DIR)/bootstrap
BOOTSTRAP_BUILD_DIR = $(BOOTSTRAP_DIR)/build
BOOTSTRAP_BUILD_STATE_DIR = $(BOOTSTRAP_DIR)/.state/build
BOOTSTRAP_LATTICE_SYSTEM_ID ?= bootstrapped
BOOTSTRAP_AWS_SYSTEM_STATE_DIR = $(BOOTSTRAP_DIR)/.state/aws/$(LATTICE_SYSTEM_ID)

LOCAL_REGISTRY = lattice-local
DEV_REGISTRY = gcr.io/lattice-dev
DEV_TAG ?= latest

KUBERNETES_BOOTSTRAP_LATTICE_IMAGE = kubernetes-bootstrap-lattice
BAZEL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE = bazel/cmd/kubernetes/bootstrap-lattice:go_image
LOCAL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE = $(LOCAL_REGISTRY)/$(KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)
DEV_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE = $(DEV_REGISTRY)/$(KUBERNETES_BOOTSTRAP_LATTICE_IMAGE):$(DEV_TAG)

KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE = kubernetes-lattice-controller-manager
BAZEL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE = bazel/cmd/kubernetes/lattice-controller-manager:go_image
LOCAL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE = $(LOCAL_REGISTRY)/$(KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)
DEV_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE = $(DEV_REGISTRY)/$(KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE):$(DEV_TAG)

CLI_IMAGE = lattice-system-cli
BAZEL_CLI_IMAGE = bazel/cmd/cli:go_image
LOCAL_CLI_IMAGE = $(LOCAL_REGISTRY)/$(CLI_IMAGE)
DEV_CLI_IMAGE = $(DEV_REGISTRY)/$(CLI_IMAGE):$(DEV_TAG)

BUILD_CONTAINER_NAME="lattice-system-builder"


# Basic build/clean/test
.PHONY: build-all
build-all: gazelle
	@bazel build //...:all

.PHONY: build-all-docker-images
build-all-docker-images: build-docker-image-cli \
						 build-docker-image-kubernetes-lattice-controller-manager \
 						 build-docker-image-kubernetes-bootstrap-lattice
	true

.PHONY: clean
clean:
	@bazel clean

.PHONY: test
test: gazelle
	@bazel test --test_output=errors //...

.PHONY: gazelle
gazelle:
	@bazel run //:gazelle

.PHONY: build-docker-image-cli
build-docker-image-cli: gazelle
	@bazel run //cmd/cli:go_image -- --norun

.PHONY: build-docker-image-kubernetes-lattice-controller-manager
build-docker-image-kubernetes-lattice-controller-manager: gazelle
	@bazel run //cmd/kubernetes/lattice-controller-manager:go_image -- --norun

.PHONY: build-docker-image-kubernetes-bootstrap-lattice
build-docker-image-kubernetes-bootstrap-lattice: gazelle
	@bazel run //cmd/kubernetes/bootstrap-lattice:go_image -- --norun


# docker build hackery
.PHONY: docker-build-all
docker-build-all: docker-build-start-build-container
	docker exec $(BUILD_CONTAINER_NAME) ./docker/wrap-ssh-creds-and-exec.sh make build-all-docker-images

.PHONY: docker-build-bazel-build
docker-build-bazel-build:
	docker build $(DIR)/docker -f $(DIR)/docker/Dockerfile.bazel-build -t lattice-build/bazel-build

.PHONY: docker-build-start-build-container
docker-build-start-build-container: docker-build-bazel-build
	$(DIR)/docker/start-build-container.sh

# docker save
.PHONY: docker-build-and-save-all
docker-build-and-save-all: docker-build-all docker-save-all

.PHONY: docker-save-all
docker-save-all:
	dest=$(dest)/$(CLI_IMAGE) make docker-save-cli
	dest=$(dest)/$(KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) make docker-save-kubernetes-lattice-controller-manager
	dest=$(dest)/$(KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) make docker-save-bootstrap-kubernetes

.PHONY: docker-save-cli
docker-save-cli: docker-tag-local-cli
	docker save $(LOCAL_CLI_IMAGE) -o $(dest)

.PHONY: docker-save-kubernetes-bootstrap-lattice
docker-save-kubernetes-bootstrap-lattice: docker-tag-local-kubernetes-bootstrap-lattice
	docker save $(LOCAL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) -o $(dest)

.PHONY: docker-save-kubernetes-lattice-controller-manager
docker-save-kubernetes-lattice-controller-manager: docker-tag-local-kubernetes-lattice-controller-manager
	docker save $(LOCAL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) -o $(dest)

.PHONY: docker-tag-local-cli
docker-tag-local-cli:
	docker tag $(BAZEL_CLI_IMAGE) $(LOCAL_CLI_IMAGE)

.PHONY: docker-tag-local-kubernetes-bootstrap-lattice
docker-tag-local-kubernetes-bootstrap-lattice:
	docker tag $(BAZEL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) $(LOCAL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-tag-local-kubernetes-lattice-controller-manager
docker-tag-local-kubernetes-lattice-controller-manager:
	docker tag $(BAZEL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) $(LOCAL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)


# docker push-dev
.PHONY: docker-push-dev-all
docker-push-dev-all: docker-push-dev-cli \
 					 docker-push-dev-kubernetes-bootstrap-lattice \
					 docker-push-dev-kubernetes-lattice-controller-manager

.PHONY: docker-build-and-push-dev-all
docker-build-and-push-dev-all: docker-build-all docker-push-dev-all

.PHONY: docker-push-dev-cli
docker-push-dev-cli: docker-tag-dev-cli
	gcloud docker -- push $(DEV_CLI_IMAGE)

.PHONY: docker-push-dev-kubernetes-bootstrap-lattice
docker-push-dev-kubernetes-bootstrap-lattice: docker-tag-dev-bootstrap-kubernetes
	gcloud docker -- push $(DEV_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-push-dev-kubernetes-lattice-controller-manager
docker-push-dev-kubernetes-lattice-controller-manager: docker-tag-dev-kubernetes-lattice-controller-manager
	gcloud docker -- push $(DEV_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)

.PHONY: docker-tag-dev-cli
docker-tag-dev-cli:
	docker tag $(BAZEL_CLI_IMAGE) $(DEV_CLI_IMAGE)

.PHONY: docker-tag-dev-kubernetes-bootstrap-lattice
docker-tag-dev-bootstrap-kubernetes:
	docker tag $(BAZEL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) $(DEV_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-tag-dev-kubernetes-lattice-controller-manager
docker-tag-dev-kubernetes-lattice-controller-manager:
	docker tag $(BAZEL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) $(DEV_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)


# cloud images
.PHONY: cloud-images-build
cloud-images-build: cloud-images-build-base-node-image cloud-images-build-master-node-image

.PHONY: cloud-images-build-base-node-image
cloud-images-build-base-node-image:
	$(BOOTSTRAP_BUILD_DIR)/build-base-node-image

.PHONY: cloud-images-build-master-node-image
cloud-images-build-master-node-image:
	$(BOOTSTRAP_BUILD_DIR)/build-master-node-image

.PHONY: cloud-images-clean
cloud-images-clean:
	rm -rf $(BOOTSTRAP_BUILD_STATE_DIR)/artifacts

.PHONY: cloud-images-clean-master-node-image
cloud-images-clean-master-node-image:
	rm -rf $(BOOTSTRAP_BUILD_STATE_DIR)/artifacts/master-node
	rm -f $(BOOTSTRAP_BUILD_STATE_DIR)/artifacts/master-node-ami-id
