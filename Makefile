.DEFAULT_GOAL=help

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

HOST_ARCH = $(shell which go >/dev/null 2>&1 && go env GOARCH)
ARCH ?= $(HOST_ARCH)

ROOT=$(shell pwd)
REPONAME = $(shell basename `pwd`)

help:  ## Display the help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build-controller
build-controller:  ## Build binddns controller
	ROOT=$(ROOT) \
	APPNAME=binddns-controller \
	BUILDDIR=controller \
	REPONAME=$(REPONAME) \
	ARCH=$(ARCH) \
	build/build.sh

.PHONY: build-webhook
build-webhook:  ## Build binddns webhook
	ROOT=$(ROOT) \
	APPNAME=binddns-webhook \
	BUILDDIR=webhook \
	REPONAME=$(REPONAME) \
	ARCH=$(ARCH) \
	build/build.sh

.PHONY: crd-generate
crd-generate:  ## Generate crd yaml to ./deploy/crd
	controller-gen paths=./pkg/apis/... crd:trivialVersions=true rbac:roleName=controller-perms output:crd:artifacts:config=deploy/crd

.PHONY: code-generate
code-generate:  ## Generate crd code
	./hack/update-codegen.sh
	controller-gen paths=./pkg/apis/... crd:trivialVersions=true rbac:roleName=controller-perms output:crd:artifacts:config=deploy/crd
