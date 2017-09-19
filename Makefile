# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
MINIKUBE_PROFILE = lattice-kubernetes-integration-dev

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

.PHONY: local-up
local-up: minikube-start local-bootstrap

.PHONY: local-down
local-down: minikube-stop

.PHONY: local-delete
local-delete: minikube-delete

.PHONY: local-bootstrap
local-bootstrap: gazelle
	@bazel run -- //cmd/bootstrap -kubeconfig ~/.kube/config -provider local
	$(DIR)/bin/seed-local-build-images.sh $(MINIKUBE_PROFILE)

.PHONY: local-clean
local-clean:
	$(DIR)/test/clean-crds.sh


.PHONY: minikube-start
minikube-start:
	@minikube start -p $(MINIKUBE_PROFILE)

.PHONY: minikube-stop
minikube-stop:
	@minikube stop -p $(MINIKUBE_PROFILE)

.PHONY: minikube-delete
minikube-delete:
	@minikube delete -p $(MINIKUBE_PROFILE)

.PHONY: minikube-ssh
minikube-ssh:
	@minikube ssh -p $(MINIKUBE_PROFILE)

.PHONY: minikube-dashboard
minikube-dashboard:
	@minikube dashboard -p $(MINIKUBE_PROFILE)
