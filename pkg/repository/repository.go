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

package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"

	"go.linka.cloud/artifact-registry/pkg/codec"
	"go.linka.cloud/artifact-registry/pkg/repository/auth"
)

type Codec = codec.Codec[Artifact]

type repoKey struct{}

func Context(ctx context.Context, r Repository) context.Context {
	return context.WithValue(ctx, repoKey{}, r)
}

func FromContext(ctx context.Context) (Repository, bool) {
	r, ok := ctx.Value(repoKey{}).(Repository)
	return r, ok
}

type StorageMiddlewareFunc = func(repoVar string) mux.MiddlewareFunc

func StorageMiddleware(ar Provider, backend string, key []byte) StorageMiddlewareFunc {
	return func(repoVar string) mux.MiddlewareFunc {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := auth.Context(r.Context(), r)
				name := mux.Vars(r)[repoVar]
				if name == "" {
					http.Error(w, "missing repository name", http.StatusBadRequest)
					return
				}
				repo, err := NewStorage(ctx, backend, name, ar, key)
				if err != nil {
					Error(w, err)
					return
				}
				defer repo.Close()
				next.ServeHTTP(w, r.WithContext(Context(ctx, repo)))
			})
		}
	}
}

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

type Provider interface {
	Index(ctx context.Context, priv string, artifacts ...Artifact) ([]Artifact, error)
	GenerateKeypair() (string, string, error)
	KeyNames() (string, string)
	Codec() Codec
	Name() string
}

type Repository interface {
	Stat(ctx context.Context, file string) (ArtifactInfo, error)
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	Write(ctx context.Context, a Artifact) error
	Delete(ctx context.Context, name string) error
	ServeFile(w http.ResponseWriter, r *http.Request, name string) error
	Key() string
	Close() error
}

var _ Artifact = (*File)(nil)

type File struct {
	name string
	data []byte
	r    io.Reader
}

func NewFile(name string, data []byte) *File {
	return &File{
		name: name,
		data: data,
		r:    bytes.NewReader(data),
	}
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.r.Read(p)
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Version() string {
	return ""
}

func (f *File) Path() string {
	return f.name
}

func (f *File) Size() int64 {
	return int64(len(f.data))
}

func (f *File) Digest() digest.Digest {
	return digest.FromBytes(f.data)
}

type info struct {
	version string
	path    string
	size    int64
	digest  digest.Digest
	meta    []byte
}

func (i *info) Name() string {
	return filepath.Base(i.path)
}

func (i *info) Version() string {
	return i.version
}

func (i *info) Path() string {
	return i.path
}

func (i *info) Size() int64 {
	return i.size
}

func (i *info) Digest() digest.Digest {
	return i.digest
}

func (i *info) Meta() []byte {
	return i.meta
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
