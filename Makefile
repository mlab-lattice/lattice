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


# Basic build/clean/test
.PHONY: build-all
build-all: gazelle
	@bazel build //...:all

.PHONY: build-all-docker-images
build-all-docker-images: build-docker-image-kubernetes-lattice-controller-manager \
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

.PHONY: build-docker-image-kubernetes-lattice-controller-manager
build-docker-image-kubernetes-lattice-controller-manager: gazelle
	@bazel run //cmd/kubernetes/lattice-controller-manager:go_image -- --norun

.PHONY: build-docker-image-kubernetes-bootstrap-lattice
build-docker-image-kubernetes-bootstrap-lattice: gazelle
	@bazel run //cmd/kubernetes/bootstrap-lattice:go_image -- --norun


# docker build hackery
.PHONY: docker-build-all
docker-build-all:
	BUILD_CMD="make build-all-docker-images" make docker-build

.PHONY: docker-build-bazel-build
docker-build-bazel-build:
	docker build $(DIR)/docker -f $(DIR)/docker/Dockerfile.bazel-build -t lattice-build/bazel-build

.PHONY: docker-build
docker-build: docker-build-bazel-build
	docker run -v $(DIR):/src -v /var/run/docker.sock:/var/run/docker.sock -v ~/.ssh/id_rsa-github:/root/.ssh/id_rsa-github lattice-build/bazel-build /src/docker/build.sh $(BUILD_CMD)


# docker save
.PHONY: docker-build-and-save-all
docker-build-and-save-all: docker-build-all docker-save-all

.PHONY: docker-save-all
docker-save-all:
	dest=$(dest)/$(KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) make docker-save-kubernetes-lattice-controller-manager
	dest=$(dest)/$(KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) make docker-save-bootstrap-kubernetes

.PHONY: docker-save-kubernetes-bootstrap-lattice
docker-save-kubernetes-bootstrap-lattice: docker-tag-local-kubernetes-bootstrap-lattice
	docker save $(LOCAL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) -o $(dest)

.PHONY: docker-save-kubernetes-lattice-controller-manager
docker-save-kubernetes-lattice-controller-manager: docker-tag-local-kubernetes-lattice-controller-manager
	docker save $(LOCAL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) -o $(dest)

.PHONY: docker-tag-local-kubernetes-bootstrap-lattice
docker-tag-local-kubernetes-bootstrap-lattice:
	docker tag $(BAZEL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) $(LOCAL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-tag-local-kubernetes-lattice-controller-manager
docker-tag-local-kubernetes-lattice-controller-manager:
	docker tag $(BAZEL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) $(LOCAL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)


# docker push-dev
.PHONY: docker-push-dev-all
docker-push-dev-all: docker-push-dev-kubernetes-bootstrap-lattice \
					 docker-push-dev-kubernetes-lattice-controller-manager

.PHONY: docker-build-and-push-dev-all
docker-build-and-push-dev-all: docker-build-all docker-push-dev-all

.PHONY: docker-push-dev-kubernetes-bootstrap-lattice
docker-push-dev-kubernetes-bootstrap-lattice: docker-tag-dev-bootstrap-kubernetes
	gcloud docker -- push $(DEV_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-push-dev-kubernetes-lattice-controller-manager
docker-push-dev-kubernetes-lattice-controller-manager: docker-tag-dev-kubernetes-lattice-controller-manager
	gcloud docker -- push $(DEV_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)

.PHONY: docker-tag-dev-kubernetes-bootstrap-lattice
docker-tag-dev-bootstrap-kubernetes:
	docker tag $(BAZEL_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE) $(DEV_KUBERNETES_BOOTSTRAP_LATTICE_IMAGE)

.PHONY: docker-tag-dev-kubernetes-lattice-controller-manager
docker-tag-dev-kubernetes-lattice-controller-manager:
	docker tag $(BAZEL_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE) $(DEV_KUBERNETES_LATTICE_CONTROLLER_MANAGER_IMAGE)


# lattice-system-cli
.PHONY: build-lattice-system-cli
build-lattice-system-cli: gazelle
	bazel build //cmd/system

.PHONY: update-local-binary-lattice-system-cli
update-local-binary-lattice-system-cli: build-lattice-system-cli
	cp -f $(DIR)/bazel-bin/cmd/system/system $(DIR)/bin/lattice-system

# provision-system
.PHONY: build-provision-system
build-provision-system: gazelle
	bazel build //cmd/provision-system

.PHONY: update-local-binary-provision-system
update-local-binary-provision-system: build-provision-system
	cp -f $(DIR)/bazel-bin/cmd/provision-system/provision-system $(DIR)/bin

# deprovision-system
.PHONY: build-deprovision-system
build-deprovision-system: gazelle
	bazel build //cmd/deprovision-system

.PHONY: update-local-binary-deprovision-system
update-local-binary-deprovision-system: build-deprovision-system
	cp -f $(DIR)/bazel-bin/cmd/deprovision-system/deprovision-system $(DIR)/bin

# cloud bootstrap
.PHONY: cloud-bootstrap-build-images
cloud-bootstrap-build-images: cloud-bootstrap-build-base-node-image cloud-bootstrap-build-master-node-image

.PHONY: cloud-bootstrap-build-base-node-image
cloud-bootstrap-build-base-node-image:
	$(BOOTSTRAP_BUILD_DIR)/build-base-node-image

.PHONY: cloud-bootstrap-build-master-node-image
cloud-bootstrap-build-master-node-image:
	$(BOOTSTRAP_BUILD_DIR)/build-master-node-image

.PHONY: cloud-bootstrap-clean-images
cloud-bootstrap-clean-images:
	rm -rf $(BOOTSTRAP_BUILD_STATE_DIR)/artifacts

.PHONY: cloud-bootstrap-clean-master-image
cloud-bootstrap-clean-master-image:
	rm -rf $(BOOTSTRAP_BUILD_STATE_DIR)/artifacts/master-node
	rm -f $(BOOTSTRAP_BUILD_STATE_DIR)/artifacts/master-node-ami-id

# aws bootstrap
.PHONY: aws-bootstrap-up
aws-bootstrap-up: cloud-bootstrap-build-images aws-bootstrap-provision-system

.PHONY: aws-bootstrap-down
aws-bootstrap-down: aws-bootstrap-deprovision-system

.PHONY: aws-bootstrap-provision-system
aws-bootstrap-provision-system:
	LATTICE_SYSTEM_ID=$(BOOTSTRAP_LATTICE_SYSTEM_ID) $(BOOTSTRAP_DIR)/scripts/aws/provision-system

.PHONY: aws-bootstrap-deprovision-system
aws-bootstrap-deprovision-system:
	LATTICE_SYSTEM_ID=$(BOOTSTRAP_LATTICE_SYSTEM_ID) $(BOOTSTRAP_DIR)/scripts/aws/deprovision-system