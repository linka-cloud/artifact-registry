MODULE = go.linka.cloud/artifact-registry

GITHUB_REPO = linka-cloud/artifact-registry

PROJECT = artifact-registry
REPOSITORY = linkacloud

UI := $(PWD)/ui

TAG = $(shell git describe --tags --exact-match 2> /dev/null)
VERSION_SUFFIX = $(shell git diff --quiet || echo "-dev")
VERSION = $(shell git describe --tags --exact-match 2> /dev/null || echo "`git describe --tags $$(git rev-list --tags --max-count=1) 2> /dev/null || echo v0.0.0`-`git rev-parse --short HEAD`")$(VERSION_SUFFIX)

COMMIT = $(shell git diff --quiet && git rev-parse --short HEAD || echo "$$(git rev-parse --short HEAD)-dirty")
COMMIT_DATE = $(shell git show -s --format=%ci $(COMMIT) 2> /dev/null || echo "not-yet")

BUILD_ARGS := -trimpath -ldflags='-s -w -X "go.linka.cloud/artifact-registry.Repo=$(GITHUB_REPO)" -X "go.linka.cloud/artifact-registry.Version=$(VERSION)" -X "go.linka.cloud/artifact-registry.Commit=$(COMMIT)" -X "go.linka.cloud/artifact-registry.Date=$(COMMIT_DATE)"'

GORELEASER_OS := $(shell [ "$$(go env GOOS)" = "darwin" ] && echo "Darwin" || echo "Linux")
GORELEASER_VERSION := v1.21.2
GORELEASER_URL := https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_$(GORELEASER_OS)_x86_64.tar.gz

show-version:
	@echo $(VERSION)

show-commit:
	@echo $(COMMIT) $(COMMIT_DATE)

BIN := $(PWD)/bin
export PATH := $(BIN):$(PATH)

bin:
	@mkdir -p $(BIN)
	@curl -sL $(GORELEASER_URL) | tar -C $(BIN) -xz goreleaser

.PHONY: tests
tests:
	@go test -v ./...

check-fmt:
	@[ "$(gofmt -l $(find . -name '*.go') 2>&1)" = "" ]

vet:
	@go list ./...|xargs go vet

build-ui:
	@yarn --cwd $(UI) install
	@yarn --cwd $(UI) build

build-go:
	@go build $(BUILD_ARGS) -o bin/lkard ./cmd/lkard
	@go build $(BUILD_ARGS) -o bin/lkar ./cmd/lkar

install: build-ui
	@go generate ./...
	@go install $(BUILD_ARGS) ./cmd/artifact-registry
	@go install $(BUILD_ARGS) ./cmd/lkar

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
build-snapshot:  bin build-ui
	@VERSION=$(VERSION) REPO=$(GITHUB_REPO) goreleaser build --snapshot --clean --parallelism 8

.PHONY: release-snapshot
release-snapshot: bin build-ui
	@VERSION=$(VERSION) REPO=$(GITHUB_REPO) goreleaser release --snapshot --clean --skip=sign,publish,announce --parallelism 8

.PHONY: build
build: bin build-ui
	@VERSION=$(VERSION) REPO=$(GITHUB_REPO) goreleaser build --clean --parallelism 8

.PHONY: release
release: bin build-ui
	@VERSION=$(VERSION) REPO=$(GITHUB_REPO) goreleaser release --clean --parallelism 8

