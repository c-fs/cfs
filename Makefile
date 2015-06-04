########################################
# cfs
########################################
.PHONY: vet
vet:
	@go vet -d ./...

.PHONY: get
get:
	@go get -d ./...

.PHONY: build
build: get
	@go build -d ./...

.PHONY: test
test: vet build
	go test -p 8 -race -d ./...
