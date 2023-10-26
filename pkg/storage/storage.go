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
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/opencontainers/go-digest"

	"go.linka.cloud/artifact-registry/pkg/codec"
	"go.linka.cloud/artifact-registry/pkg/slices"
)

var ErrInvalidArtifactType = errors.New("invalid image's artifact type")

type Codec = codec.Codec[Artifact]

// TODO(adphi): keep only the read closer interface,
// move to ArtifactInfo to its own interface and add a method to retrieve the ArtifactInfo, e.g. Stat() ArtifactInfo

type Artifact interface {
	io.ReadCloser
	// Name is the name of the artifact, e.g. "jq".
	Name() string
	// Path is the path of the artifact in the repository.
	Path() string
	// Arch is the architecture of the artifact.
	Arch() string
	// Version is the version of the artifact.
	Version() string
	// Size is the binary size of the artifact.
	Size() int64
	Digest() digest.Digest
}

type ArtifactInfo interface {
	Name() string
	Version() string
	Arch() string
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
	Init(ctx context.Context) error
	Stat(ctx context.Context, file string) (ArtifactInfo, error)
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	Write(ctx context.Context, a Artifact) error
	Delete(ctx context.Context, name string) error
	Artifacts(ctx context.Context) ([]Artifact, error)
	ServeFile(w http.ResponseWriter, r *http.Request, name string) error
	Size(ctx context.Context) (int64, error)
	Key() string
	Close() error
}

func As[T Artifact](as []Artifact) ([]T, error) {
	return slices.MapErr(as, func(v Artifact) (T, error) {
		var z T
		pkg, ok := v.(T)
		if !ok {
			return z, fmt.Errorf("invalid artifact type %T", v)
		}
		return pkg, nil
	})
}

func MustAs[T Artifact](as []Artifact) []T {
	return must(As[T](as))
}

func AsArtifact[T Artifact](as []T) []Artifact {
	return slices.Map(as, func(v T) Artifact {
		return v
	})
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
