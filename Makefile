# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CLOUD_IMAGE_DIR = $(DIR)/cloud-images
CLOUD_IMAGE_BUILD_DIR = $(CLOUD_IMAGE_DIR)/build
CLOUD_IMAGE_BUILD_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/build
CLOUD_IMAGE_AWS_SYSTEM_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/aws/$(LATTICE_SYSTEM_ID)

OS := $(shell uname)
USER := $(shell whoami)

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

.PHONY: check
check: gazelle format vet lint-no-export-comments

.PHONY: format
format:
	@gofmt -w .
	@terraform fmt .

.PHONY: lint
lint: install-golint
	@golint ./...

.PHONY: lint-no-export-comments
lint-no-export-comments: install-golint
	@golint ./... | grep -v " or be unexported"

.PHONY: install-golint
install-golint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then go get github.com/golang/lint/golint; fi

.PHONY: vet
vet: install-govet
	@go tool vet .

.PHONY: install-govet
install-govet:
	@go tool vet 2>/dev/null; if [ $$? -eq 3 ]; then go get golang.org/x/tools/cmd/vet; fi


# docker
.PHONY: docker-push-image-stable
docker-push-image-stable:
	@if [ $(OS) != Linux ]; then echo "Must run docker-push-image on Linux" && exit 1; fi
	bazel run //docker:push-stable-$(IMAGE)
	bazel run //docker:push-stable-debug-$(IMAGE)

.PHONY: docker-push-image-user
docker-push-image-user:
	@if [ $(OS) != Linux ]; then echo "Must run docker-push-image on Linux" && exit 1; fi
	bazel run //docker:push-user-$(IMAGE)
	bazel run //docker:push-user-debug-$(IMAGE)

.PHONY: docker-push-all-images-stable
docker-push-all-images-stable:
	make docker-push-image-stable IMAGE=envoy-prepare
	make docker-push-image-stable IMAGE=kubernetes-bootstrap-lattice
	make docker-push-image-stable IMAGE=kubernetes-component-builder
	make docker-push-image-stable IMAGE=kubernetes-envoy-xds-api-rest-per-node
	make docker-push-image-stable IMAGE=kubernetes-lattice-controller-manager
	make docker-push-image-stable IMAGE=kubernetes-manager-api-rest

.PHONY: docker-push-all-images-user
docker-push-all-images-user:
	make docker-push-image-user IMAGE=envoy-prepare
	make docker-push-image-user IMAGE=kubernetes-bootstrap-lattice
	make docker-push-image-user IMAGE=kubernetes-component-builder
	make docker-push-image-user IMAGE=kubernetes-envoy-xds-api-rest-per-node
	make docker-push-image-user IMAGE=kubernetes-lattice-controller-manager
	make docker-push-image-user IMAGE=kubernetes-manager-api-rest

# local binaries
.PHONY: update-binaries
update-binaries: update-binary-cli-admin update-binary-cli-user

.PHONY: update-binary-cli-admin
update-binary-cli-admin: build
	@bazel build //cmd/cli/admin
	cp -f $(DIR)/bazel-bin/cmd/cli/admin/admin $(DIR)/bin/lattice-admin

.PHONY: update-binary-cli-user
update-binary-cli-user: build
	@bazel build //cmd/cli/user
	cp -f $(DIR)/bazel-bin/cmd/cli/user/user $(DIR)/bin/lattice-system

# docker build hackery
.PHONY: docker-hack-enter-build-shell
docker-hack-enter-build-shell: docker-hack-build-start-build-container
	docker exec -it -e USER=$(USER) $(CONTAINER_NAME_BUILD) ./docker/bazel-builder/wrap-creds-and-exec.sh /bin/bash

.PHONY: docker-hack-push-image
docker-hack-push-image: docker-hack-build-start-build-container
	docker exec -e USER=$(USER) $(CONTAINER_NAME_BUILD) ./docker/bazel-builder/wrap-creds-and-exec.sh make docker-push-image IMAGE=$(IMAGE)

.PHONY: docker-hack-push-all-images-stable
docker-hack-push-all-images-stable: docker-hack-build-start-build-container
	docker exec -e USER=$(USER) $(CONTAINER_NAME_BUILD) ./docker/bazel-builder/wrap-creds-and-exec.sh make docker-push-all-images-stable IMAGE=$(IMAGE)

.PHONY: docker-hack-push-all-images-user
docker-hack-push-all-images-user: docker-hack-build-start-build-container
	docker exec -e USER=$(USER) $(CONTAINER_NAME_BUILD) ./docker/bazel-builder/wrap-creds-and-exec.sh make docker-push-all-images-user IMAGE=$(IMAGE)

.PHONY: docker-hack-build-bazel-build
docker-hack-build-bazel-build:
	docker build --build-arg user=$(USER) $(DIR)/docker -f $(DIR)/docker/bazel-builder/Dockerfile.bazel-build -t lattice-build/bazel-build

.PHONY: docker-hack-build-start-build-container
docker-hack-build-start-build-container: docker-hack-build-bazel-build
	USER=$(USER) $(DIR)/docker/bazel-builder/start-build-container.sh

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
