# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CLOUD_IMAGE_DIR = $(DIR)/cloud-images
CLOUD_IMAGE_BUILD_DIR = $(CLOUD_IMAGE_DIR)/build
CLOUD_IMAGE_BUILD_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/build
CLOUD_IMAGE_AWS_SYSTEM_STATE_DIR = $(CLOUD_IMAGE_DIR)/.state/aws/$(LATTICE_SYSTEM_ID)

OS := $(shell uname)
USER := $(shell whoami)

# build/clean
.PHONY: build
build: gazelle
	@bazel build //...:all

.PHONY: build.all
build.all: build.darwin \
           build.linux

.PHONY: build.darwin
build.darwin: gazelle
	@bazel build --cpu darwin //...:all

.PHONY: build.linux
build.linux: gazelle
	@bazel build --cpu k8 //...:all

.PHONY: gazelle
gazelle:
	@bazel run //:gazelle

.PHONY: clean
clean:
	@bazel clean


# testing
.PHONY: test
test: gazelle
	@bazel test --test_output=errors //pkg/...

.PHONY: test.no-cache
test.no-cache: gazelle
	@bazel test --cache_test_results=no --test_output=errors //pkg/...

.PHONY: test.verbose
test.verbose: gazelle
	@bazel test --test_output=all --test_env -v //pkg/...


# e2e testing
.PHONY: e2e-test
e2e-test: e2e-test.build
	@$(DIR)/bazel-bin/test/e2e/darwin_amd64_stripped/go_default_test -cluster-url $(CLUSTER_URL)

.PHONY: e2e-test.provider
e2e-test.provider: e2e-test.build
	@$(DIR)/bazel-bin/test/e2e/darwin_amd64_stripped/go_default_test -cloud-provider $(PROVIDER)

.PHONY: e2e-test.local
e2e-test.local: e2e-test.build
	@$(MAKE) e2e-test.provider PROVIDER=local

.PHONY: e2e-test.build
e2e-test.build: gazelle
	@bazel build //test/e2e/...


# formatting/linting
.PHONY: check
check: gazelle \
       format  \
       vet     \
       lint-no-export-comments

.PHONY: format
format:
	@gofmt -w .
	@terraform fmt .

.PHONY: lint
lint: install.golint
	@golint ./... | grep -v "customresource/generated" | grep -v "zz_generated."

.PHONY: lint-no-export-comments
lint-no-export-comments: install.golint
	@$(MAKE) lint | grep -v " or be unexported" | grep -v "comment on exported "

.PHONY: vet
vet: install.govet
	@go tool vet .


# tool installation
.PHONY: install.golint
install.golint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then go get github.com/golang/lint/golint; fi

.PHONY: install.govet
install.govet:
	@go tool vet 2>/dev/null; if [ $$? -eq 3 ]; then go get golang.org/x/tools/cmd/vet; fi


# git
.PHONY: git.install-hooks
git.install-hooks:
	cp -f scripts/git/pre-commit.sh .git/hooks/pre-commit
	cp -f scripts/git/pre-push.sh .git/hooks/pre-push


# docker
.PHONY: docker.push-image-stable
docker.push-image-stable: gazelle
	bazel run --cpu k8 //docker:push-stable-$(IMAGE)
	bazel run --cpu k8 //docker:push-stable-debug-$(IMAGE)

.PHONY: docker.push-image-user
docker.push-image-user: gazelle
	bazel run --cpu k8 //docker:push-user-$(IMAGE)
	bazel run --cpu k8 //docker:push-user-debug-$(IMAGE)

DOCKER_IMAGES := envoy-prepare                                 \
                 kubernetes-api-server-rest                    \
                 kubernetes-aws-master-node-attach-etcd-volume \
                 kubernetes-aws-master-node-register-dns       \
                 kubernetes-component-builder                  \
                 kubernetes-envoy-xds-api-rest-per-node        \
                 kubernetes-lattice-controller-manager         \
                 kubernetes-local-dns-controller               \
                 latticectl

STABLE_CONTAINER_PUSHES := $(addprefix docker.push-image-stable-,$(DOCKER_IMAGES))
USER_CONTAINER_PUSHES := $(addprefix docker.push-image-user-,$(DOCKER_IMAGES))

.PHONY: $(STABLE_CONTAINER_PUSHES)
$(STABLE_CONTAINER_PUSHES):
	@$(MAKE) docker.push-image-stable IMAGE=$(patsubst docker.push-image-stable-%,%,$@)

.PHONY: $(USER_CONTAINER_PUSHES)
$(USER_CONTAINER_PUSHES):
	@$(MAKE) docker.push-image-user IMAGE=$(patsubst docker.push-image-user-%,%,$@)

.PHONY: docker.push-all-stable
docker.push-all-stable:
	@for image in $(DOCKER_IMAGES); do \
		$(MAKE) docker.push-image-stable-$$image ; \
	done

.PHONY: docker.push-all-user
docker.push-all-user:
	@for image in $(DOCKER_IMAGES); do \
		$(MAKE) docker.push-image-user-$$image ; \
	done


# binaries
.PHONY: binary.update-latticectl
binary.update-latticectl: binary.update-latticectl-darwin-amd64 \
                          binary.update-latticectl-linux-amd64

.PHONY: binary.update-latticectl-darwin-amd64
binary.update-latticectl-darwin-amd64: gazelle
	@bazel build --cpu darwin //cmd/latticectl
	cp -f $(DIR)/bazel-bin/cmd/latticectl/darwin_amd64_stripped/latticectl $(DIR)/bin/latticectl-darwin-amd64

.PHONY: binary.update-latticectl-linux-amd64
binary.update-latticectl-linux-amd64: gazelle
	@bazel build --cpu k8 //cmd/latticectl
	cp -f $(DIR)/bazel-bin/cmd/latticectl/linux_amd64_pure_stripped/latticectl $(DIR)/bin/latticectl-linux-amd64


# kubernetes
.PHONY: kubernetes.update-dependencies
kubernetes.update-dependencies:
	LATTICE_ROOT=$(DIR) KUBERNETES_VERSION=$(VERSION) $(DIR)/scripts/kubernetes/dependencies/update-kubernetes-version.sh
	$(MAKE) kubernetes.regenerate-custom-resource-clients VERSION=$(VERSION)

.PHONY: kubernetes.regenerate-custom-resource-clients
kubernetes.regenerate-custom-resource-clients:
	KUBERNETES_VERSION=$(VERSION) $(DIR)/scripts/kubernetes/codegen/regenerate.sh


# cloud images
.PHONY: cloud-images.build
cloud-images.build: cloud-images.build-base-node-image \
                    cloud-images.build-master-node-image

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
