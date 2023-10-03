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

FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN go build -trimpath -ldflags="-s -w" -o artifact-registry ./cmd/artifact-registry

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=0 /app/artifact-registry /usr/local/bin/artifact-registry

ENTRYPOINT ["/usr/local/bin/artifact-registry"]