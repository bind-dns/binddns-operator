.DEFAULT_GOAL=build

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

HOST_ARCH = $(shell which go >/dev/null 2>&1 && go env GOARCH)
ARCH ?= $(HOST_ARCH)

ROOT=$(shell pwd)
APPNAME = $(shell basename `pwd`)

# build
.PHONY: build
build:
	ROOT=$(ROOT) \
	APPNAME=$(APPNAME) \
	ARCH=$(ARCH) \
	build/build.sh

# crd-generate
.PHONY: crd-generate
crd-generate:
	controller-gen paths=./pkg/apis/... crd:trivialVersions=true rbac:roleName=controller-perms output:crd:artifacts:config=deploy/crd

# code-generate
.PHONY: code-generate
code-generate:
	./hack/update-codegen.sh
	controller-gen paths=./pkg/apis/... crd:trivialVersions=true rbac:roleName=controller-perms output:crd:artifacts:config=deploy/crd
