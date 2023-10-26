MODULE = go.linka.cloud/artifact-registry

GITHUB_REPO = linka-cloud/artifact-registry

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
	@go install -trimpath -ldflags "-s -w -X '$(MODULE).Version=$(VERSION)' -X '$(MODULE).BuildDate=$(shell date -Iseconds)'" ./cmd/artifact-registry
	@go install -trimpath -ldflags "-s -w -X '$(MODULE).Version=$(VERSION)' -X '$(MODULE).BuildDate=$(shell date -Iseconds)'" ./cmd/lkar

DOCKER_BUILDX_ARGS := build --pull --load --build-arg VERSION=$(VERSION)

docker: docker-build docker-scan docker-push

docker-scan:
	@trivy image --severity "HIGH,CRITICAL" --exit-code 100 $(REPOSITORY)/$(PROJECT):$(VERSION)

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

.PHONY: completions
completions:
	@rm -rf completions
	@mkdir -p completions
	@for shell in bash zsh fish powershell; do \
		go run ./cmd/lkar completion $$shell > completions/lkar.$$shell; \
	done

.PHONY: cli-docs
cli-docs:
	@rm -rf ./docs/{lkar,lkard}
	@go run -tags=docs ./cmd/lkar docs ./docs/lkar
	@go run -tags=docs ./cmd/lkard docs ./docs/lkard

PHONY: build-snapshot
build-snapshot:  build-ui
	@VERSION=$(VERSION) REPO=$(GITHUB_REPO) goreleaser build --snapshot --clean --parallelism 8

.PHONY: release-snapshot
release-snapshot: build-ui
	@VERSION=$(VERSION) REPO=$(GITHUB_REPO) goreleaser release --snapshot --clean --skip=sign,publish,announce --parallelism 8

.PHONY: build
build: build-ui
	@VERSION=$(VERSION) REPO=$(GITHUB_REPO) goreleaser build --clean --parallelism 8

.PHONY: release
release: build-ui
	@VERSION=$(VERSION) REPO=$(GITHUB_REPO) goreleaser release --clean --parallelism 8

