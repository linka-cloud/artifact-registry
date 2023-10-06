IMAGE := "linkacloud/artifact-registry"
MODULE = go.linka.cloud/artifact-registry

install: docker-build
	@go generate ./...
	@go install -trimpath -ldflags "-s -w -X '$(MODULE).Version=$(VERSION)' -X '$(MODULE).BuildDate=$(shell date)'" ./cmd/artifact-registry
	@go install -trimpath -ldflags "-s -w -X '$(MODULE).Version=$(VERSION)' -X '$(MODULE).BuildDate=$(shell date)'" ./cmd/lkar

docker: docker-build docker-push

.PHONY: docker-build
docker-build:
	@docker build -t $(IMAGE) .

.PHONY: docker-push
docker-push:
	@docker push $(IMAGE)
