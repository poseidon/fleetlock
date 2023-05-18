export CGO_ENABLED:=0

DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION=$(shell git describe --tags --match=v* --always --dirty)

REPO=github.com/poseidon/fleetlock
LOCAL_REPO=poseidon/fleetlock
IMAGE_REPO=quay.io/poseidon/fleetlock

LD_FLAGS="-w -X main.version=$(VERSION)"

.PHONY: all
all: build test vet fmt

.PHONY: build
build:
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

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: image
image: \
	image-amd64 \
	image-arm64

image-%:
	buildah bud -f Dockerfile \
	-t $(LOCAL_REPO):$(VERSION)-$* \
	--arch $* --override-arch $* \
	--format=docker .

push-%:
	buildah tag $(LOCAL_REPO):$(VERSION)-$* $(IMAGE_REPO):$(VERSION)-$*
	buildah push --format v2s2 $(IMAGE_REPO):$(VERSION)-$*

manifest:
	buildah manifest create $(IMAGE_REPO):$(VERSION)
	buildah manifest add $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)-amd64
	buildah manifest add --variant v8 $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)-arm64
	buildah manifest inspect $(IMAGE_REPO):$(VERSION)
	buildah manifest push -f v2s2 $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)

lock:
	curl -H "fleet-lock-protocol: true" -d @examples/body.json http://127.0.0.1:8080/v1/pre-reboot

unlock:
	curl -H "fleet-lock-protocol: true" -d @examples/body.json http://127.0.0.1:8080/v1/steady-state


