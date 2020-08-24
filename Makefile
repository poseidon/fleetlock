export CGO_ENABLED:=0
export GO111MODULE=on
export GOFLAGS=-mod=vendor

DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION=$(shell git describe --tags --match=v* --always --abbrev=0 --dirty)

REPO=github.com/poseidon/fleetlock
LOCAL_REPO=poseidon/fleetlock
IMAGE_REPO=quay.io/poseidon/fleetlock

LD_FLAGS="-w -X $(REPO)/cmd.version=$(VERSION)"

.PHONY: all
all: build test vet lint fmt

.PHONY: build
build: bin/fleetlock

.PHONY: bin/fleetlock
bin/fleetlock:
	@go build -o bin/fleetlock -ldflags $(LD_FLAGS) $(REPO)/cmd/fleetlock

.PHONY: test
test:
	@go test ./... -cover

.PHONY: vet
vet:
	@go vet -all ./...

.PHONY: lint
lint:
	@golint -set_exit_status `go list ./...`

.PHONY: fmt
fmt:
	@test -z $$(go fmt ./...)

.PHONY: image
image:
	buildah bud -t $(LOCAL_REPO):$(VERSION) .
	buildah tag $(LOCAL_REPO):$(VERSION) $(LOCAL_REPO):latest

.PHONY: push
push:
	buildah tag $(LOCAL_REPO):$(VERSION) $(IMAGE_REPO):$(VERSION)
	buildah tag $(LOCAL_REPO):$(VERSION) $(IMAGE_REPO):latest
	buildah push docker://$(IMAGE_REPO):$(VERSION)
	buildah push docker://$(IMAGE_REPO):latest

.PHONY: update
update:
	@GOFLAGS="" go get -u ./...
	@go mod tidy

.PHONY: vendor
vendor:
	@go mod vendor

lock:
	curl -H "fleet-lock-protocol: true" -d @examples/body.json http://127.0.0.1:8080/v1/pre-reboot

unlock:
	curl -H "fleet-lock-protocol: true" -d @examples/body.json http://127.0.0.1:8080/v1/steady-state


