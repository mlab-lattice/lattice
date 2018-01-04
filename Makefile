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

.PHONY: test-verbose
test-verbose: gazelle
	@bazel test --test_output=all --test_env -v  //...

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


# kubernetes
.PHONY: kubernetes.update-dependencies
kubernetes.update-dependencies:
	LATTICE_ROOT=$(DIR) KUBERNETES_VERSION=$(VERSION) $(DIR)/scripts/k8s/dependencies/update-kubernetes-version.sh
	make kubernetes.regenerate-custom-resource-clients VERSION=$(VERSION)

.PHONY: kubernetes.regenerate-custom-resource-clients
kubernetes.regenerate-custom-resource-clients:
	KUBERNETES_VERSION=$(VERSION) $(DIR)/scripts/k8s/codegen/regenerate.sh


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
	make docker.push-image-stable IMAGE=kubernetes-component-builder
	make docker.push-image-stable IMAGE=kubernetes-envoy-xds-api-rest-per-node
	make docker.push-image-stable IMAGE=kubernetes-lattice-controller-manager
	make docker.push-image-stable IMAGE=kubernetes-manager-api-rest
	make docker.push-image-stable IMAGE=kubernetes-local-dns
	make docker.push-image-stable IMAGE=lattice-cli-admin

.PHONY: docker.push-all-images-user
docker.push-all-images-user:
	make docker.push-image-user IMAGE=envoy-prepare
	make docker.push-image-user IMAGE=kubernetes-component-builder
	make docker.push-image-user IMAGE=kubernetes-envoy-xds-api-rest-per-node
	make docker.push-image-user IMAGE=kubernetes-lattice-controller-manager
	make docker.push-image-user IMAGE=kubernetes-manager-api-rest
	make docker.push-image-user IMAGE=kubernetes-local-dns
	make docker.push-image-user IMAGE=lattice-cli-admin


# binaries
.PHONY: update-binaries
update-binaries: update-binary-cli-admin update-binary-cli-user

.PHONY: update-binary-cli-admin
update-binary-cli-admin: update-binary-cli-admin-darwin update-binary-cli-admin-linux

.PHONY: update-binary-cli-user
update-binary-cli-user: update-binary-cli-user-darwin update-binary-cli-user-linux

.PHONY: update-binary-cli-admin-darwin
update-binary-cli-admin-darwin: gazelle
	@bazel build --cpu darwin //cmd/cli/admin
	cp -f $(DIR)/bazel-bin/cmd/cli/admin/darwin_amd64_stripped/admin $(DIR)/bin/lattice-admin-darwin-amd64

.PHONY: update-binary-cli-admin-linux
update-binary-cli-admin-linux: gazelle
	@bazel build --cpu k8 //cmd/cli/admin
	cp -f $(DIR)/bazel-bin/cmd/cli/admin/linux_amd64_pure_stripped/admin $(DIR)/bin/lattice-admin-linux-amd64

.PHONY: update-binary-cli-user-darwin
update-binary-cli-user-darwin: gazelle
	@bazel build --cpu darwin //cmd/cli/user
	cp -f $(DIR)/bazel-bin/cmd/cli/user/darwin_amd64_stripped/user $(DIR)/bin/lattice-user-darwin-amd64

.PHONY: update-binary-cli-user-linux
update-binary-cli-user-linux: gazelle
	@bazel build --cpu k8 //cmd/cli/user
	cp -f $(DIR)/bazel-bin/cmd/cli/user/linux_amd64_pure_stripped/user $(DIR)/bin/lattice-user-linux-amd64


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
