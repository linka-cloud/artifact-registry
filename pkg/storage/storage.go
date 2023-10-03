// Copyright 2023 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/opencontainers/go-digest"

	"go.linka.cloud/artifact-registry/pkg/codec"
)

type Codec = codec.Codec[Artifact]

type Artifact interface {
	io.Reader
	Name() string
	Version() string
	Path() string
	Size() int64
	Digest() digest.Digest
}

type ArtifactInfo interface {
	Name() string
	Version() string
	Path() string
	Size() int64
	Digest() digest.Digest
	Meta() []byte
}

type Repository interface {
	Index(ctx context.Context, priv string, artifacts ...Artifact) ([]Artifact, error)
	GenerateKeypair() (string, string, error)
	KeyNames() (string, string)
	Codec() Codec
	Name() string
}

type Storage interface {
	Stat(ctx context.Context, file string) (ArtifactInfo, error)
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	Write(ctx context.Context, a Artifact) error
	Delete(ctx context.Context, name string) error
	ServeFile(w http.ResponseWriter, r *http.Request, name string) error
	Key() string
	Close() error
}

func As[T Artifact](as []Artifact) ([]T, error) {
	var packages []T
	for _, v := range as {
		pkg, ok := v.(T)
		if !ok {
			return nil, fmt.Errorf("invalid artifact type %T", v)
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}

func MustAs[T Artifact](as []Artifact) []T {
	packages, err := As[T](as)
	if err != nil {
		panic(err)
	}
	return packages
}
