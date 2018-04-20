# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

# build
.PHONY: build
build: gazelle
	@bazel build \
		//pkg/...:all \
		//cmd/...:all \
		//test/...:all

.PHONY: build.all
build.all: build.darwin \
           build.linux

.PHONY: build.darwin
build.darwin: gazelle
	@bazel build \
		--platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 \
		//pkg/...:all \
		//cmd/...:all \
		//test/...:all

.PHONY: build.linux
build.linux: gazelle
	@bazel build \
		--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//pkg/...:all \
		//cmd/...:all \
		//test/...:all

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
DOCKER_IMAGES := kubernetes-api-server-rest             \
                 kubernetes-component-builder           \
                 kubernetes-envoy-prepare               \
                 kubernetes-envoy-xds-api-rest-per-node \
                 kubernetes-lattice-controller-manager  \
                 kubernetes-local-dns-controller        \
                 latticectl

.PHONY: docker.push-image
docker.push-image: gazelle \
                   docker.push-image-no-gazelle

.PHONY: docker.push-image-no-gazelle
docker.push-image-no-gazelle:
	# currently only pushing debug images
	@bazel run \
		--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		--workspace_status_command "REGISTRY=$(REGISTRY) CHANNEL=$(CHANNEL) $(DIR)/scripts/bazel/docker-workspace-status.sh" \
		//docker:push-debug-$(IMAGE)

IMAGE_PUSHES := $(addprefix docker.push-image-,$(DOCKER_IMAGES))
.PHONY: $(IMAGE_PUSHES)
$(IMAGE_PUSHES):
	@$(MAKE) docker.push-image IMAGE=$(patsubst docker.push-image-%,%,$@)

IMAGE_PUSHES_NO_GAZELLE := $(addprefix docker.push-image-no-gazelle-,$(DOCKER_IMAGES))
.PHONY: $(IMAGE_PUSHES_NO_GAZELLE)
$(IMAGE_PUSHES_NO_GAZELLE):
	@$(MAKE) docker.push-image-no-gazelle IMAGE=$(patsubst docker.push-image-no-gazelle-%,%,$@)

.PHONY: docker.push-all
docker.push-all: gazelle
	@for image in $(DOCKER_IMAGES); do \
		$(MAKE) docker.push-image-no-gazelle-$$image || exit 1; \
	done

.PHONY: docker.save-image
docker.save-image: gazelle
	@bazel run \
		--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//docker:debug-$(IMAGE) \
		-- --norun

IMAGE_SAVES := $(addprefix docker.save-image-,$(DOCKER_IMAGES))

.PHONY: $(IMAGE_SAVES)
$(IMAGE_SAVES):
	@$(MAKE) docker.save-image IMAGE=$(patsubst docker.save-image-%,%,$@)

.PHONY: docker.run-shell
docker.run-shell: docker.save-image
	@docker run -it --entrypoint sh bazel/docker:debug-$(IMAGE)

IMAGE_RUNS := $(addprefix docker.run-shell-,$(DOCKER_IMAGES))

.PHONY: $(IMAGE_RUNS)
$(IMAGE_RUNS):
	@$(MAKE) docker.run-shell IMAGE=$(patsubst docker.run-shell-%,%,$@)

# kubernetes
.PHONY: kubernetes.update-dependencies
kubernetes.update-dependencies:
	LATTICE_ROOT=$(DIR) KUBERNETES_VERSION=$(VERSION) $(DIR)/scripts/kubernetes/dependencies/update-kubernetes-version.sh
	$(MAKE) kubernetes.regenerate-custom-resource-clients VERSION=$(VERSION)

.PHONY: kubernetes.regenerate-custom-resource-clients
kubernetes.regenerate-custom-resource-clients:
	KUBERNETES_VERSION=$(VERSION) $(DIR)/scripts/kubernetes/codegen/regenerate.sh
