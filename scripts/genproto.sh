#!/bin/bash -e
#
# Generate all cfs protobuf bindings.
# Run from repository root.
#

SHA="16256d3ce6929458613798ee44b7914a3f59f5c6"

if ! protoc --version > /dev/null; then
	echo "could not find protoc, is it installed + in PATH?"
	echo "check https://github.com/google/protobuf/ about how to install"
	# TODO: use release version
	echo "please install protoc at commit 42809ef8fef9e4d76267eb21bcb8a856f10ba418"
	exit 255
fi

# Ensure we have the right version of protoc-gen-go by building it every time.
export GOPATH=${PWD}/proto/gopath
go get github.com/golang/protobuf/protoc-gen-go
pushd ${GOPATH}/src/github.com/golang/protobuf/
	git reset --hard ${SHA}
	make
popd

export PATH="${GOPATH}/bin:${PATH}"

pushd ./proto
	protoc --go_out=plugins=grpc:. *.proto
popd

rm -rf ${GOPATH}
