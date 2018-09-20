# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

# build
.PHONY: build
build: gazelle \
       build.no-gazelle

.PHONY: build.no-gazelle
build.no-gazelle: PLATFORM ?=
build.no-gazelle: TARGET ?= //cmd/... //pkg/...
build.no-gazelle:
	@bazel build \
		$(PLATFORM) \
		$(TARGET)

.PHONY: build.platform.all
build.platform.all: build.platform.darwin \
                    build.platform.linux

.PHONY: build.platform.darwin
build.platform.darwin: gazelle
	@$(MAKE) build.no-gazelle PLATFORM=--platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64

.PHONY: build.platform.linux
build.platform.linux: gazelle
	@$(MAKE) build.no-gazelle PLATFORM=--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64

.PHONY: gazelle
gazelle:
	@bazel run //:gazelle

.PHONY: clean
clean:
	@bazel clean


# testing
.PHONY: test
test: TARGET ?= //pkg/...
test: OUTPUT ?= errors
test: ARGS ?=
test: gazelle
	@bazel test \
		$(ARGS) \
		--test_output=$(OUTPUT) \
		$(TARGET)

.PHONY: test.no-cache
test.no-cache:
	@$(MAKE) test ARGS=--cache_test_results=no

.PHONY: test.verbose
test.verbose:
	@$(MAKE) test OUTPUT=all ARGS="--test_env -v"


# formatting/linting
.PHONY: check
check: gazelle \
       format  \
       vet     \
       lint.no-export-comments

.PHONY: format
format:
	@gofmt -w .
	@terraform fmt .

.PHONY: lint
lint: install.golint
	@golint ./... | grep -v "customresource/generated" | grep -v "zz_generated."

.PHONY: lint.no-export-comments
lint.no-export-comments: install.golint
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
	cp -f hack/git/pre-commit.sh $(DIR)/.git/hooks/pre-commit
	cp -f hack/git/pre-push.sh $(DIR)/.git/hooks/pre-push


# docker
DOCKER_IMAGES := kubernetes-api-server-rest             \
                 kubernetes-container-builder           \
                 kubernetes-envoy-prepare               \
                 kubernetes-envoy-xds-api-grpc-per-node \
                 kubernetes-lattice-controller-manager  \
                 kubernetes-local-dns-controller        \
                 latticectl                             \
                 mock-api-server

.PHONY: docker.push
docker.push: gazelle \
             docker.push-no-gazelle

.PHONY: docker.push-no-gazelle
docker.push-no-gazelle:
	@bazel run \
		--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		--workspace_status_command "REGISTRY=$(REGISTRY) CHANNEL=$(CHANNEL) $(DIR)/hack/bazel/docker-workspace-status.sh" \
		//docker:push-$(IMAGE)

IMAGE_PUSHES := $(addprefix docker.push.,$(DOCKER_IMAGES))
.PHONY: $(IMAGE_PUSHES)
$(IMAGE_PUSHES):
	@$(MAKE) docker.push IMAGE=$(patsubst docker.push.%,%,$@)

IMAGE_PUSHES_NO_GAZELLE := $(addprefix docker.push-no-gazelle.,$(DOCKER_IMAGES))
.PHONY: $(IMAGE_PUSHES_NO_GAZELLE)
$(IMAGE_PUSHES_NO_GAZELLE):
	@$(MAKE) docker.push-no-gazelle IMAGE=$(patsubst docker.push-no-gazelle.%,%,$@)

.PHONY: docker.push.all
docker.push.all: gazelle
	@for image in $(DOCKER_IMAGES); do \
		$(MAKE) docker.push-no-gazelle.$$image || exit 1; \
	done

.PHONY: docker.push-stripped
docker.push-stripped: gazelle \
                      docker.push-stripped-no-gazelle

.PHONY: docker.push-stripped-no-gazelle
docker.push-stripped-no-gazelle:
	@bazel run \
		--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		--workspace_status_command "REGISTRY=$(REGISTRY) CHANNEL=$(CHANNEL) $(DIR)/hack/bazel/docker-workspace-status.sh" \
		//docker:push-$(IMAGE)-stripped

STRIPPED_IMAGE_PUSHES := $(addprefix docker.push-stripped.,$(DOCKER_IMAGES))
.PHONY: $(STRIPPED_IMAGE_PUSHES)
$(STRIPPED_IMAGE_PUSHES):
	@$(MAKE) docker.push-stripped IMAGE=$(patsubst docker.push-stripped.%,%,$@)

STRIPPED_IMAGE_PUSHES_NO_GAZELLE := $(addprefix docker.push-stripped-no-gazelle.,$(DOCKER_IMAGES))
.PHONY: $(STRIPPED_IMAGE_PUSHES_NO_GAZELLE)
$(STRIPPED_IMAGE_PUSHES_NO_GAZELLE):
	@$(MAKE) docker.push-stripped-no-gazelle IMAGE=$(patsubst docker.push-stripped-no-gazelle.%,%,$@)

.PHONY: docker.push-stripped.all
docker.push-stripped.all: gazelle
	@for image in $(DOCKER_IMAGES); do \
		$(MAKE) docker.push-stripped-no-gazelle.$$image || exit 1; \
	done

.PHONY: docker.save
docker.save: gazelle
	@bazel run \
		--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//docker:$(IMAGE) \
		-- --norun

IMAGE_SAVES := $(addprefix docker.save.,$(DOCKER_IMAGES))

.PHONY: $(IMAGE_SAVES)
$(IMAGE_SAVES):
	@$(MAKE) docker.save IMAGE=$(patsubst docker.save.%,%,$@)

.PHONY: docker.run
docker.run: docker.save
	@docker run -it --entrypoint sh bazel/docker:$(IMAGE)

IMAGE_RUNS := $(addprefix docker.run.,$(DOCKER_IMAGES))

.PHONY: $(IMAGE_RUNS)
$(IMAGE_RUNS):
	@$(MAKE) docker.run IMAGE=$(patsubst docker.run.%,%,$@)

# kubernetes
.PHONY: kubernetes.update-dependencies
kubernetes.update-dependencies:
	LATTICE_ROOT=$(DIR) KUBERNETES_VERSION=$(VERSION) $(DIR)/hack/kubernetes/dependencies/update-kubernetes-version.sh
	$(MAKE) kubernetes.regenerate-custom-resource-clients VERSION=$(VERSION)

.PHONY: kubernetes.regenerate-custom-resource-clients
kubernetes.regenerate-custom-resource-clients:
	KUBERNETES_VERSION=$(VERSION) $(DIR)/hack/kubernetes/codegen/regenerate.sh
	@$(MAKE) gazelle
