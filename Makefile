export CGO_ENABLED:=0
export GO111MODULE=on

DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION=$(shell git describe --tags --match=v* --always --dirty)

REPO=github.com/poseidon/fleetlock
LOCAL_REPO=poseidon/fleetlock
IMAGE_REPO=quay.io/poseidon/fleetlock

LD_FLAGS="-w -X main.version=$(VERSION)"

.PHONY: all
all: build test vet fmt

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

.PHONY: fmt
fmt:
	@test -z $$(go fmt ./...)

.PHONY: image
image: \
	image-amd64 \
	image-arm64

image-%:
	buildah bud -f Dockerfile \
	-t $(LOCAL_REPO):$(VERSION)-$* \
	--arch $* --override-arch $* \
	--format=docker .

lock:
	curl -H "fleet-lock-protocol: true" -d @examples/body.json http://127.0.0.1:8080/v1/pre-reboot

unlock:
	curl -H "fleet-lock-protocol: true" -d @examples/body.json http://127.0.0.1:8080/v1/steady-state


