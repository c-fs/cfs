########################################
# cfs
########################################

ORG_PATH:=github.com/c-fs
REPO_PATH:=$(ORG_PATH)/cfs

$(eval TEST := $(shell cd ../../.. && find $(REPO_PATH) -name '*_test.go' |  xargs -L 1 dirname | uniq | sort))

.PHONY: build
# TODO

.PHONY: test
test: go-vet build
	go test -p 8 -race $(TEST)

.PHONY: go-vet
go-vet:
	@find . -name '*.go' | xargs -L 1 go tool vet

.PHONY: proto
proto:
	protoc -I proto proto/*.proto --go_out=plugins=grpc:proto
