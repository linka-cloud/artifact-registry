MODULE = go.linka.cloud/artifact-registry

GITHUB_REPO = linka-cloud/artifact-registry

PROJECT = artifact-registry
REPOSITORY = linkacloud

UI := $(PWD)/ui

TAG = $(shell git describe --tags --exact-match 2> /dev/null)
VERSION_SUFFIX = $(shell git diff --quiet || echo "-dev")
VERSION = $(shell git describe --tags --exact-match 2> /dev/null || echo "`git describe --tags $$(git rev-list --tags --max-count=1) 2> /dev/null || echo v0.0.0`-`git rev-parse --short HEAD`")$(VERSION_SUFFIX)

CHART_TAG = $(shell git describe --tags --match="helm/*" HEAD 2>/dev/null | sed 's|helm/||')
ifneq ($(CHART_TAG),)
	CHART_VERSION = $(CHART_TAG)
else
	CHART_VERSION = "0.0.0"
endif

COMMIT = $(shell git diff --quiet && git rev-parse --short HEAD || echo "$$(git rev-parse --short HEAD)-dirty")
COMMIT_DATE = $(shell git show -s --format=%ci $(COMMIT) 2> /dev/null || echo "not-yet")

BUILD_ARGS := -trimpath -ldflags='-s -w -X "go.linka.cloud/artifact-registry.Repo=$(GITHUB_REPO)" -X "go.linka.cloud/artifact-registry.Version=$(VERSION)" -X "go.linka.cloud/artifact-registry.Commit=$(COMMIT)" -X "go.linka.cloud/artifact-registry.Date=$(COMMIT_DATE)"'

OS=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)

GORELEASER_VERSION := v1.21.2
GORELEASER_URL := https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_$(shell uname -s)_$(shell uname -m).tar.gz

HELM_VERSION := v3.13.1
HELM_URL := https://get.helm.sh/helm-$(HELM_VERSION)-$(OS)-$(ARCH).tar.gz

show-os:
	@echo $(OS)

show-version:
	@echo $(VERSION)

show-commit:
	@echo $(COMMIT) $(COMMIT_DATE)

show-chart-version:
	@echo $(CHART_VERSION)

BIN := $(PWD)/bin
export PATH := $(BIN):$(PATH)

bin:
	@mkdir -p $(BIN)
	@curl -sL $(GORELEASER_URL) | tar -C $(BIN) -xz goreleaser
	@curl -sL $(HELM_URL) | tar -C $(BIN) -xz --strip-components 1 "$(OS)-$(ARCH)/helm"
	@helm plugin list | grep unittest 2>&1 >/dev/null || helm plugin install https://github.com/helm-unittest/helm-unittest.git

.PHONY: tests
tests:
	@go test -v ./...

.PHONY: helm-version
helm-version:
	@sed helm/artifact-registry/Chart.in.yaml -e 's|{{ CHART_VERSION }}|$(CHART_VERSION)|' -e 's|{{ APP_VERSION }}|$(TAG)|' > helm/artifact-registry/Chart.yaml

helm-tests: bin helm-version
	@helm unittest helm/artifact-registry

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

.PHONY: helm
helm: helm-version
	@mkdir -p dist
	@helm package -d dist helm/artifact-registry

helm-release: helm
ifneq ($(TAG),)
	@curl --user "$(REPO_USER):$(REPO_PASSWORD)" --upload-file dist/artifact-registry-$(CHART_VERSION).tgz https://helm.linka.cloud/push
endif

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

