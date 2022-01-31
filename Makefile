#!/usr/bin/make -f

PACKAGES=$(shell go list ./...)
BUILDDIR ?= $(CURDIR)/build
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf


test:
	go test ./...

install:
	go install ./cmd/celestia

## build: Build DALC binary.
build:
	@echo "--> Building DALC"
	@go build ./cmd/dalc
.PHONY: build

# Build DALC docker image
docker-build:
	@docker build --platform linux/amd64 -f docker/Dockerfile -t ghcr.io/celestiaorg/dalc:latest .

proto-all: proto-gen proto-lint
.PHONY: proto-all

proto-gen:
	@docker pull -q tendermintdev/docker-build-proto
	@echo "Generating Protobuf files"
	@docker run -v $(shell pwd):/workspace --workdir /workspace tendermintdev/docker-build-proto sh ./scripts/protocgen.sh
.PHONY: proto-gen
proto-lint:
	@$(DOCKER_BUF) lint --error-format=json
