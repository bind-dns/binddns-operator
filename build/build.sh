#!/bin/bash

set -eo pipefail

# set bin output
mkdir -p "${ROOT}"/build/bin/"${ARCH}"/

# set ldflags variables.
REVISION=$(git rev-parse --short HEAD)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
TAG=$(git branch | head -n 1 | awk '{print $4}' | head -c-2)
GOVERSION=$(go version)
BUILDTIME=$(date "+%Y-%m-%d %H:%M:%S")

# set package name.
PACKAGE_APPNAME="github.com/bind-dns/${APPNAME}"

LDFLAGS="-s -w -X '${PACKAGE_APPNAME}/version.APPNAME=${APPNAME}' -X '${PACKAGE_APPNAME}/version.REVISION=${REVISION}' -X '${PACKAGE_APPNAME}/version.BRANCH=${BRANCH}' -X '${PACKAGE_APPNAME}/version.TAG=${TAG}' -X '${PACKAGE_APPNAME}/version.GOVERSION=${GOVERSION}' -X '${PACKAGE_APPNAME}/version.BUILDTIME=${BUILDTIME}' -X '${PACKAGE_APPNAME}/version.BINDVERSION=${BINDVERSION}'"

# go build
CGO_ENABLED=0 go build -mod vendor -v -x -ldflags "${LDFLAGS}" -o ${ROOT}/build/binddns-controller/bin/${ARCH}/${APPNAME} ${ROOT}/cmd/controller/main.go

# docker build
docker build -t binddns-operator:latest ${ROOT}/build/binddns-controller
