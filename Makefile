# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CLOUD_IMAGE_DIR = $(DIR)/cloud-images
CLOUD_IMAGE_BUILD_DIR = $(CLOUD_IMAGE_DIR)/build
CLOUD_IMAGE_BUILD_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/build
CLOUD_IMAGE_AWS_SYSTEM_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/aws/$(LATTICE_SYSTEM_ID)

OS := $(shell uname)
USER := $(shell whoami)

# Basic build/clean/test
.PHONY: build
build: gazelle
	@bazel build //...:all

.PHONY: build-linux
build-linux: gazelle
	@bazel build --cpu k8 //...:all

.PHONY: build-all
build-all: build build-linux

.PHONY: clean
clean:
	@bazel clean

.PHONY: test
test: gazelle
	@bazel test --test_output=errors //...

.PHONY: gazelle
gazelle:
	@bazel run //:gazelle

.PHONY: check
check: gazelle format vet lint-no-export-comments

.PHONY: format
format:
	@gofmt -w .
	@terraform fmt .

.PHONY: lint
lint: install.golint
	@golint ./...

.PHONY: lint-no-export-comments
lint-no-export-comments: install.golint
	@golint ./... | grep -v " or be unexported"

.PHONY: install.golint
install.golint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then go get github.com/golang/lint/golint; fi

.PHONY: vet
vet: install.govet
	@go tool vet .

.PHONY: install.govet
install.govet:
	@go tool vet 2>/dev/null; if [ $$? -eq 3 ]; then go get golang.org/x/tools/cmd/vet; fi


# git
.PHONY: git.install-hooks
git.install-hooks:
	cp -f scripts/git/pre-commit.sh .git/hooks/pre-commit
	cp -f scripts/git/pre-push.sh .git/hooks/pre-push


# kubernetes
.PHONY: kubernetes.update-dependencies
kubernetes.update-dependencies:
	LATTICE_ROOT=$(DIR) KUBERNETES_VERSION=$(VERSION) $(DIR)/scripts/kubernetes/dependencies/update-kubernetes-version.sh
	make kubernetes.regenerate-custom-resource-clients VERSION=$(VERSION)

.PHONY: kubernetes.regenerate-custom-resource-clients
kubernetes.regenerate-custom-resource-clients:
	KUBERNETES_VERSION=$(VERSION) $(DIR)/scripts/kubernetes/codegen/regenerate.sh


# docker
.PHONY: docker.push-image-stable
docker.push-image-stable:
	bazel run --cpu k8 //docker:push-stable-$(IMAGE)
	bazel run --cpu k8 //docker:push-stable-debug-$(IMAGE)

.PHONY: docker.push-image-user
docker.push-image-user:
	bazel run --cpu k8 //docker:push-user-$(IMAGE)
	bazel run --cpu k8 //docker:push-user-debug-$(IMAGE)

.PHONY: docker.push-all-images-stable
docker.push-all-images-stable:
	make docker.push-image-stable IMAGE=envoy-prepare
	make docker.push-image-stable IMAGE=kubernetes-aws-master-node-attach-etcd-volume
	make docker.push-image-stable IMAGE=kubernetes-aws-master-node-register-dns
	make docker.push-image-stable IMAGE=kubernetes-component-builder
	make docker.push-image-stable IMAGE=kubernetes-envoy-xds-api-rest-per-node
	make docker.push-image-stable IMAGE=kubernetes-lattice-controller-manager
	make docker.push-image-stable IMAGE=kubernetes-manager-api-rest
	make docker.push-image-stable IMAGE=latticectl

.PHONY: docker.push-all-images-user
docker.push-all-images-user:
	make docker.push-image-user IMAGE=envoy-prepare
	make docker.push-image-user IMAGE=kubernetes-aws-master-node-attach-etcd-volume
	make docker.push-image-user IMAGE=kubernetes-aws-master-node-register-dns
	make docker.push-image-user IMAGE=kubernetes-component-builder
	make docker.push-image-user IMAGE=kubernetes-envoy-xds-api-rest-per-node
	make docker.push-image-user IMAGE=kubernetes-lattice-controller-manager
	make docker.push-image-user IMAGE=kubernetes-manager-api-rest
	make docker.push-image-user IMAGE=latticectl


# binaries
.PHONY: binary.update-latticectl
binary.update-latticectl: binary.update-latticectl-darwin-amd64 binary.update-latticectl-linux-amd64

.PHONY: binary.update-latticectl-darwin-amd64
binary.update-latticectl-darwin-amd64: gazelle
	@bazel build --cpu darwin //cmd/latticectl
	cp -f $(DIR)/bazel-bin/cmd/latticectl/darwin_amd64_stripped/latticectl $(DIR)/bin/latticectl-darwin-amd64

.PHONY: binary.update-latticectl-linux-amd64
binary.update-latticectl-linux-amd64: gazelle
	@bazel build --cpu k8 //cmd/latticectl
	cp -f $(DIR)/bazel-bin/cmd/latticectl/linux_amd64_pure_stripped/latticectl $(DIR)/bin/latticectl-linux-amd64


# cloud images
.PHONY: cloud-images.build
cloud-images.build: cloud-images.build-base-node-image cloud-images.build-master-node-image

.PHONY: cloud-images.build-base-node-image
cloud-images.build-base-node-image:
	$(CLOUD_IMAGE_BUILD_DIR)/build-base-node-image

.PHONY: cloud-images.build-master-node-image
cloud-images.build-master-node-image:
	$(CLOUD_IMAGE_BUILD_DIR)/build-master-node-image

.PHONY: cloud-images.clean
cloud-images.clean:
	rm -rf $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts

.PHONY: cloud-images.clean-master-node-image
cloud-images.clean-master-node-image:
	rm -rf $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts/master-node
	rm -f $(CLOUD_IMAGE_BUILD_STATE_DIR)/artifacts/master-node-ami-id
