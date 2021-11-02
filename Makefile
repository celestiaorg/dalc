#!/usr/bin/make -f

PACKAGES=$(shell go list ./...)
BUILDDIR ?= $(CURDIR)/build
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-all: proto-gen proto-lint
.PHONY: proto-all

proto-gen:
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen sh ./scripts/protocgen.sh

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json