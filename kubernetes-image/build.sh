#!/bin/bash

ORG_PATH=github.com/c-fs
REPO_PATH=$ORG_PATH/cfs

go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"' -o cfs ${REPO_PATH}/server
go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"' -o cfsctl ${REPO_PATH}/cfsctl
docker build -t yunxing/cfs:k8s .
rm cfs cfsctl
