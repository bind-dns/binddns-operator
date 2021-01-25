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
PACKAGE="github.com/bind-dns/${REPONAME}"

LDFLAGS="-s -w -X '${PACKAGE}/version.APPNAME=${APPNAME}' -X '${PACKAGE}/version.REVISION=${REVISION}' -X '${PACKAGE}/version.BRANCH=${BRANCH}' -X '${PACKAGE}/version.TAG=${TAG}' -X '${PACKAGE}/version.GOVERSION=${GOVERSION}' -X '${PACKAGE}/version.BUILDTIME=${BUILDTIME}' -X '${PACKAGE}/version.BINDVERSION=${BINDVERSION}'"

# go build
CGO_ENABLED=0 go build -mod vendor -v -x -ldflags "${LDFLAGS}" -o ${ROOT}/build/${APPNAME}/bin/${ARCH}/${APPNAME} ${ROOT}/cmd/${BUILDDIR}/main.go

# copy webapp
cp -r ${ROOT}/webapp ${ROOT}/build/${APPNAME}

# docker build
docker build -t ${APPNAME}:latest ${ROOT}/build/${APPNAME}

# delete webapp
rm -rf ${ROOT}/build/${APPNAME}/webapp
