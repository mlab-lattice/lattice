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

CLI_IMAGE = lattice-system-cli
BAZEL_CLI_IMAGE = bazel/cmd/cli:go_image
LOCAL_CLI_IMAGE = $(LOCAL_REGISTRY)/$(CLI_IMAGE)
DEV_CLI_IMAGE = $(DEV_REGISTRY)/$(CLI_IMAGE):$(DEV_TAG)

KUBERNETES_BOOTSTRAP_LATTICE_IMAGE = kubernetes-bootstrap-lattice
BAZEL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE = bazel/cmd/kubernetes/bootstrap-lattice:go_image
LOCAL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE = $(LOCAL_REGISTRY)/$(KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)
DEV_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE = $(DEV_REGISTRY)/$(KUBERNETES_BOOTSTRAP_LATTICE_IMAGE):$(DEV_TAG)

KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE = kubernetes-system-manager-api-rest
BAZEL_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE = bazel/cmd/envoy/xds-api/rest/per-node:go_image
LOCAL_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE = $(LOCAL_REGISTRY)/$(KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE)
DEV_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE = $(DEV_REGISTRY)/$(KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE):$(DEV_TAG)

KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE = kubernetes-lattice-controller-manager
BAZEL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE = bazel/cmd/kubernetes/lattice-controller-manager:go_image
LOCAL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE = $(LOCAL_REGISTRY)/$(KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)
DEV_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE = $(DEV_REGISTRY)/$(KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE):$(DEV_TAG)

KUBERNETES_MANAGER_API_REST_IMAGE = kubernetes-system-manager-api-rest
BAZEL_KUBERNETES_MANAGER_API_REST_IMAGE = bazel/cmd/manager/api-rest-kubernetes:go_image
LOCAL_KUBERNETES_MANAGER_API_REST_IMAGE = $(LOCAL_REGISTRY)/$(KUBERNETES_MANAGER_API_REST_IMAGE)
DEV_KUBERNETES_MANAGER_API_REST_IMAGE_IMAGE = $(DEV_REGISTRY)/$(KUBERNETES_MANAGER_API_REST_IMAGE):$(DEV_TAG)

BUILD_CONTAINER_NAME="lattice-system-builder"


# Basic build/clean/test
.PHONY: build
build: gazelle
	@bazel build //...:all

.PHONY: build-docker-images
build-docker-images: build-docker-image-cli \
 					 build-docker-image-kubernetes-bootstrap-lattice \
					 build-docker-image-kubernetes-lattice-controller-manager \
					 build-docker-image-kubernetes-manager-api-rest
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

.PHONY: build-docker-image-kubernetes-bootstrap-lattice
build-docker-image-kubernetes-bootstrap-lattice: gazelle
	@bazel run //cmd/kubernetes/bootstrap-lattice:go_image -- --norun

.PHONY: build-docker-image-kubernetes-envoy-xds-api-rest-per-node
build-docker-image-kubernetes-envoy-xds-api-rest-per-node: gazelle
	@bazel run //cmd/kubernetes/envoy/xds-api/rest/per-node:go_image -- --norun

.PHONY: build-docker-image-kubernetes-lattice-controller-manager
build-docker-image-kubernetes-lattice-controller-manager: gazelle
	@bazel run //cmd/kubernetes/lattice-controller-manager:go_image -- --norun

.PHONY: build-docker-image-kubernetes-manager-api-rest
build-docker-image-kubernetes-manager-api-rest: gazelle
	@bazel run //cmd/kubernetes/manager/api-rest:go_image -- --norun


# docker build hackery
.PHONY: docker-build
docker-build: docker-build-start-build-container
	docker exec $(BUILD_CONTAINER_NAME) ./docker/wrap-ssh-creds-and-exec.sh make build-docker-images

.PHONY: docker-build-bazel-build
docker-build-bazel-build:
	docker build $(DIR)/docker -f $(DIR)/docker/Dockerfile.bazel-build -t lattice-build/bazel-build

.PHONY: docker-build-start-build-container
docker-build-start-build-container: docker-build-bazel-build
	$(DIR)/docker/start-build-container.sh

# docker save
.PHONY: docker-build-and-save
docker-build-and-save: docker-build docker-save

.PHONY: docker-save
docker-save: docker-save-cli \
			 docker-save-kubernetes-bootstrap-lattice \
			 docker-save-kubernetes-envoy-xds-api-rest-per-node \
			 docker-save-kubernetes-lattice-controller-manager

.PHONY: docker-save-cli
docker-save-cli: docker-tag-local-cli
	docker save $(LOCAL_CLI_IMAGE) -o $(dest)/$(CLI_IMAGE)

.PHONY: docker-save-kubernetes-bootstrap-lattice
docker-save-kubernetes-bootstrap-lattice: docker-tag-local-kubernetes-bootstrap-lattice
	docker save $(LOCAL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) -o $(dest)/$(KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-save-kubernetes-envoy-xds-api-rest-per-node
docker-save-kubernetes-envoy-xds-api-rest-per-node: docker-tag-local-kubernetes-envoy-xds-api-rest-per-node
	docker save $(LOCAL_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE) -o $(dest)/$(KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE)

.PHONY: docker-save-kubernetes-lattice-controller-manager
docker-save-kubernetes-lattice-controller-manager: docker-tag-local-kubernetes-lattice-controller-manager
	docker save $(LOCAL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) -o $(dest)/$(KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)

.PHONY: docker-save-kubernetes-manager-api-rest
docker-save-kubernetes-manager-api-rest: docker-tag-local-kubernetes-lattice-controller-manager
	docker save $(LOCAL_KUBERNETES_MANAGER_API_REST_IMAGE) -o $(dest)/$(KUBERNETES_MANAGER_API_REST_IMAGE)

.PHONY: docker-tag-local-cli
docker-tag-local-cli:
	docker tag $(BAZEL_CLI_IMAGE) $(LOCAL_CLI_IMAGE)

.PHONY: docker-tag-local-kubernetes-bootstrap-lattice
docker-tag-local-kubernetes-bootstrap-lattice:
	docker tag $(BAZEL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) $(LOCAL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-tag-local-kubernetes-envoy-xds-api-rest-per-node
docker-tag-local-kubernetes-envoy-xds-api-rest-per-node:
	docker tag $(BAZEL_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE) $(LOCAL_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE)

.PHONY: docker-tag-local-kubernetes-lattice-controller-manager
docker-tag-local-kubernetes-lattice-controller-manager:
	docker tag $(BAZEL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) $(LOCAL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)

.PHONY: docker-tag-local-kubernetes-manager-api-rest
docker-tag-local-kubernetes-manager-api-rest:
	docker tag $(BAZEL_KUBERNETES_MANAGER_API_REST_IMAGE) $(LOCAL_KUBERNETES_MANAGER_API_REST_IMAGE)


# docker push-dev
.PHONY: docker-push-dev
docker-push-dev: docker-push-dev-cli \
 				 docker-push-dev-kubernetes-bootstrap-lattice \
 				 docker-push-dev-kubernetes-envoy-xds-api-rest-per-node \
				 docker-push-dev-kubernetes-lattice-controller-manager

.PHONY: docker-build-and-push-dev
docker-build-and-push-dev: docker-build docker-push-dev

.PHONY: docker-push-dev-cli
docker-push-dev-cli: docker-tag-dev-cli
	gcloud docker -- push $(DEV_CLI_IMAGE)

.PHONY: docker-push-dev-kubernetes-bootstrap-lattice
docker-push-dev-kubernetes-bootstrap-lattice: docker-tag-dev-kubernetes-bootstrap-lattice
	gcloud docker -- push $(DEV_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-push-dev-kubernetes-envoy-xds-api-rest-per-node
docker-push-dev-kubernetes-envoy-xds-api-rest-per-node: docker-tag-dev-kubernetes-envoy-xds-api-rest-per-node
	gcloud docker -- push $(DEV_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE)

.PHONY: docker-push-dev-kubernetes-lattice-controller-manager
docker-push-dev-kubernetes-lattice-controller-manager: docker-tag-dev-kubernetes-lattice-controller-manager
	gcloud docker -- push $(DEV_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)

.PHONY: docker-push-dev-kubernetes-manager-api-rest
docker-push-dev-kubernetes-manager-api-rest: docker-tag-dev-kubernetes-manager-api-rest
	gcloud docker -- push $(DEV_KUBERNETES_MANAGER_API_REST_IMAGE)

.PHONY: docker-tag-dev-cli
docker-tag-dev-cli:
	docker tag $(BAZEL_CLI_IMAGE) $(DEV_CLI_IMAGE)

.PHONY: docker-tag-dev-kubernetes-bootstrap-lattice
docker-tag-dev-kubernetes-bootstrap-lattice:
	docker tag $(BAZEL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) $(DEV_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-tag-dev-kubernetes-envoy-xds-api-rest-per-node
docker-tag-dev-kubernetes-envoy-xds-api-rest-per-node:
	docker tag $(BAZEL_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE) $(DEV_KUBERNETES_ENVOY_XDS_API_REST_PER_NODE_IMAGE)

.PHONY: docker-tag-dev-kubernetes-lattice-controller-manager
docker-tag-dev-kubernetes-lattice-controller-manager:
	docker tag $(BAZEL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) $(DEV_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)

.PHONY: docker-tag-dev-kubernetes-manager-api-rest
docker-tag-dev-kubernetes-manager-api-rest:
	docker tag $(BAZEL_KUBERNETES_MANAGER_API_REST_IMAGE) $(DEV_KUBERNETES_MANAGER_API_REST_IMAGE)


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
