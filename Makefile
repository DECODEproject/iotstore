# Template for Go apps

# The binary to build (just the basename)
BIN := iotstore

# The projects root import path (under GOPATH)
PKG := github.com/DECODEproject/iotstore

# Docker Hub ID to which docker images should be pushed
REGISTRY ?= decodeproject

# Default architecture to build
ARCH ?= amd64

# Version string - to be added to the binary
VERSION := $(shell git describe --tags --always --dirty)

# Build date - to be added to the binary
BUILD_DATE := $(shell date -u "+%FT%H:%M:%S%Z")

# Do not change the following variables

PWD := $(shell pwd)

SRC_DIRS := cmd pkg

ALL_ARCH := amd64 arm arm64

ifeq ($(ARCH),amd64)
	BASE_IMAGE?=alpine
	BUILD_IMAGE?=golang:1.12-alpine
endif
ifeq ($(ARCH),arm)
	BASE_IMAGE?=arm32v7/busybox
	BUILD_IMAGE?=arm32v7/golang:1.12-stretch
endif
ifeq ($(ARCH),arm64)
	BASE_IMAGE?=arm64v8/busybox
	BUILD_IMAGE?=arm64v8/golang:1.12-alpine
endif

IMAGE := $(REGISTRY)/$(BIN)-$(ARCH)
all: build

build-%:
	@$(MAKE) --no-print-directory ARCH=$* build

container-%:
	@$(MAKE) --no-print-directory ARCH=$* container

push-%:
	@$(MAKE) --no-print-directory ARCH=$* push

all-build: $(addprefix build-, $(ALL_ARCH))

all-container: $(addprefix container-, $(ALL_ARCH))

all-push: $(addprefix push-, $(ALL_ARCH))

build: bin/$(ARCH)/$(BIN) ## Build our binary inside a container

bin/$(ARCH)/$(BIN): .build-dirs .compose
	@echo "--> Building in the containerized environment"
	@docker-compose -f .docker-compose-$(ARCH).yml build
	@docker-compose -f .docker-compose-$(ARCH).yml \
		run \
		--rm \
		-u $$(id -u):$$(id -g) \
		--no-deps \
		app \
		/bin/sh -c " \
			ARCH=$(ARCH) \
			VERSION=$(VERSION) \
			PKG=$(PKG) \
			BUILD_DATE=$(BUILD_DATE) \
			BINARY_NAME=$(BIN) \
			./build/build.sh \
		"

shell: .shell-$(ARCH) ## Open shell in containerized environment
.shell-$(ARCH): .build-dirs .compose
	@echo "--> Launching shell in the containerized environment"
	@docker-compose -f .docker-compose-$(ARCH).yml \
		run \
		--rm \
		-u "$$(id -u):$$(id -g)" \
		app \
		/bin/sh -c " \
			./build/dev.sh\
		"

.PHONY: test
test: .build-dirs .compose ## Run tests in the containerized environment
	@echo "--> Running tests in the containerized environment"
	@docker-compose -f .docker-compose-$(ARCH).yml \
		run \
		--rm \
		-u $$(id -u):$$(id -g) \
		-e "IOTSTORE_DATABASE_URL=postgres://iotstore:password@postgres/iotstore_test?sslmode=disable" \
		app \
		/bin/sh -c " \
			./build/test.sh $(SRC_DIRS) \
		"

DOTFILE_IMAGE = $(subst :,_,$(subst /,_,$(IMAGE))-$(VERSION))

container: .container-$(DOTFILE_IMAGE) container-name ## Create delivery container image
.container-$(DOTFILE_IMAGE): bin/$(ARCH)/$(BIN) Dockerfile.in
	@sed \
		-e 's|ARG_BIN|$(BIN)|g' \
		-e 's|ARG_ARCH|$(ARCH)|g' \
		-e 's|ARG_FROM|$(BASE_IMAGE)|g' \
		Dockerfile.in > .dockerfile-in-$(ARCH)
	@docker build -t $(IMAGE):$(VERSION) -f .dockerfile-in-$(ARCH) .
	@docker images -q $(IMAGE):$(VERSION) > $@

.PHONY: container-name
container-name: ## Show the name of the delivery container
	@echo "  container: $(IMAGE):$(VERSION)"

.PHONY: .compose
.compose: ## Create environment specific compose file
	@sed \
		-e 's|ARG_FROM|$(BUILD_IMAGE)|g' \
		-e 's|ARG_WORKDIR|/go/src/$(PKG)|g' \
		Dockerfile.dev > .dockerfile-dev-$(ARCH)
	@sed \
		-e 's|ARG_DOCKERFILE|.dockerfile-dev-$(ARCH)|g' \
		-e 's|ARG_IMAGE|$(IMAGE)-dev:$(VERSION)|g' \
		-e 's|ARG_PWD|$(PWD)|g' \
		-e 's|ARG_PKG|$(PKG)|g' \
		-e 's|ARG_ARCH|$(ARCH)|g' \
		-e 's|ARG_BIN|$(BIN)|g' \
		docker-compose.yml > .docker-compose-$(ARCH).yml

.PHONY: .build-dirs
.build-dirs: ## creates build directories
	@mkdir -p bin/$(ARCH)
	@mkdir -p .go/src/$(PKG) .go/pkg .go/bin .go/std/$(ARCH) .cache/go-build .coverage

.PHONY: version
version: ## returns the current version
	@echo Version: $(VERSION) - $(BUILD_DATE) $(IMAGE)

.PHONY: push
push: .push-$(DOTFILE_IMAGE) push-name
.push-$(DOTFILE_IMAGE):
	@docker push $(IMAGE):$(VERSION)
	@docker images -q $(IMAGE):$(VERSION) > $@

.PHONY: push-name
push-name:
	@echo "  pushed $(IMAGE):$(VERSION)"

.PHONY: start
start: .compose ## start compose services
	@docker-compose -f .docker-compose-$(ARCH).yml \
		up

.PHONY: teardown
teardown: .compose ## teardown compose services
	@docker-compose -f .docker-compose-$(ARCH).yml \
		down -v

.PHONY: clean
clean: container-clean bin-clean ## remove all artefacts

.PHONY: container-clean
container-clean: ## clean container artefacts
	rm -rf .container-* .dockerfile-* .docker-compose-* .push-*

.PHONY: bin-clean
bin-clean: ## remove generated build artefacts
	rm -rf .go bin .cache .coverage

.PHONY: psql
psql: ## open psql shell
	@docker-compose -f .docker-compose-$(ARCH).yml start postgres
	@sleep 1
	@docker exec -it iotstore_postgres_1 psql -U iotstore iotstore_development
