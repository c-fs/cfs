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

.PHONY: docker
docker:
	# Static binary built here may fail to call os/user/lookup functions due to library
	# conflict. (http://stackoverflow.com/questions/8140439/why-would-it-be-impossible-to-fully-statically-link-an-application)
	# Because cfs doesn't use these functions, it is ok to ignore the error.
	go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"' -o cfs ${REPO_PATH}/server
	docker build -t c-fs/cfs .
