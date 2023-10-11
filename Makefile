MODULE = go.linka.cloud/artifact-registry

PROJECT = artifact-registry
REPOSITORY = linkacloud

UI := $(PWD)/ui

TAG = $(shell git describe --tags --exact-match 2> /dev/null)
VERSION_SUFFIX = $(shell git diff --quiet || echo "-dev")
VERSION = $(shell git describe --tags --exact-match 2> /dev/null || echo "`git describe --tags $$(git rev-list --tags --max-count=1) 2> /dev/null || echo v0.0.0`-`git rev-parse --short HEAD`")$(VERSION_SUFFIX)
show-version:
	@echo $(VERSION)

build-ui:
	@yarn --cwd $(UI) install
	@yarn --cwd $(UI) build

install: build-ui
	@go generate ./...
	@go install -trimpath -ldflags "-s -w -X '$(MODULE).Version=$(VERSION)' -X '$(MODULE).BuildDate=$(shell date)'" ./cmd/artifact-registry
	@go install -trimpath -ldflags "-s -w -X '$(MODULE).Version=$(VERSION)' -X '$(MODULE).BuildDate=$(shell date)'" ./cmd/lkar

DOCKER_BUILDX_ARGS := build --pull --load

docker: docker-build docker-push

.PHONY: docker-build
docker-build:
	@docker buildx $(DOCKER_BUILDX_ARGS) -t $(REPOSITORY)/$(PROJECT):$(VERSION) -t $(REPOSITORY)/$(PROJECT):dev .
ifneq ($(TAG),)
	@docker image tag $(REPOSITORY)/$(PROJECT):$(VERSION) $(REPOSITORY)/$(PROJECT):latest
endif

.PHONY: docker-push
docker-push:
	@docker image push $(REPOSITORY)/$(PROJECT):$(VERSION)
	@docker image push $(REPOSITORY)/$(PROJECT):dev
ifneq ($(TAG),)
	@docker image push $(REPOSITORY)/$(PROJECT):latest
endif
