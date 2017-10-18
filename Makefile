# https://stackoverflow.com/questions/18136918/how-to-get-current-relative-directory-of-your-makefile
DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
MINIKUBE_PROFILE = lattice-kubernetes-integration-dev
BOOTSTRAP_DIR = $(DIR)/bootstrap
BOOTSTRAP_BUILD_DIR = $(BOOTSTRAP_DIR)/build
BOOTSTRAP_BUILD_STATE_DIR = $(BOOTSTRAP_DIR)/.state/build
BOOTSTRAP_LATTICE_SYSTEM_ID ?= bootstrapped
BOOTSTRAP_AWS_SYSTEM_STATE_DIR = $(BOOTSTRAP_DIR)/.state/aws/$(LATTICE_SYSTEM_ID)

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

# minikube
.PHONY: minikube-start
minikube-start:
	@minikube start -p $(MINIKUBE_PROFILE) --kubernetes-version v1.8.0 --bootstrapper kubeadm --extra-config=kubelet.resolv-conf=$(DIR)/tmp/local-k8s-resolv.conf

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

# local on top of minikube
.PHONY: local-up
local-up: minikube-start local-bootstrap

.PHONY: local-down
local-down: minikube-stop

.PHONY: local-delete
local-delete: minikube-delete

.PHONY: local-bootstrap
local-bootstrap: gazelle
	$(DIR)/bin/seed-local-images.sh $(MINIKUBE_PROFILE)
	@bazel run -- //cmd/bootstrap -kubeconfig ~/.kube/config -provider local

.PHONY: local-clean
local-clean:
	$(DIR)/test/clean-crds.sh

.PHONY: run-controller
run-controller: gazelle
	bazel run -- //cmd/controller-manager -kubeconfig ~/.kube/config -v 5 -logtostderr -provider local

.PHONY: seed-rollout
seed-rollout: gazelle
	bazel run -- //test/system-build-and-rollout -kubeconfig ~/.kube/config -v 5 -logtostderr

.PHONY: local-reset
local-reset: local-delete local-up seed-rollout run-controller
	true

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