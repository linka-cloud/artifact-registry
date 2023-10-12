# Copyright 2023 Linka Cloud  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM node:alpine as react-builder

WORKDIR /app

COPY ui/package.json ui/yarn.lock ./

RUN yarn install --frozen-lockfile

COPY ui/ ./

RUN yarn build


FROM golang:alpine as go-builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg
COPY version.go ./version.go
COPY ui/ui.go ./ui/ui.go

COPY --from=react-builder /app/build ./ui/build

ARG VERSION=dev

RUN go build -trimpath -ldflags="-s -w -X go.linka.cloud/artifact-registry.Version=${VERSION} -X go.linka.cloud/artifact-registry.BuildDate=$(date -Iseconds)" -o lkard ./cmd/lkard
RUN go build -trimpath -ldflags="-s -w -X go.linka.cloud/artifact-registry.Version=${VERSION} -X go.linka.cloud/artifact-registry.BuildDate=$(date -Iseconds)" -o lkar ./cmd/lkar

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=go-builder /app/lkard /usr/local/bin/lkard
COPY --from=go-builder /app/lkar /usr/local/bin/lkar

ENTRYPOINT ["/usr/local/bin/lkard"]
