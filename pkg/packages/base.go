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

package packages

import (
	"io"
	"net/http"

	"go.linka.cloud/artifact-registry/pkg/storage"
)

type ArtifactFactory func(r *http.Request, reader io.Reader, size int64, key string) (storage.Artifact, error)

type Base struct{}

func Upload(fn ArtifactFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var (
			reader io.ReadCloser
			size   int64
		)
		if file, header, err := r.FormFile("file"); err == nil {
			reader, size = file, header.Size
		} else {
			reader, size = r.Body, r.ContentLength
		}
		defer reader.Close()
		s, ok := storage.FromContext(ctx)
		if !ok {
			http.Error(w, "missing storage in context", http.StatusInternalServerError)
			return
		}
		pkg, err := fn(r, reader, size, s.Key())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := s.Write(ctx, pkg); err != nil {
			storage.Error(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func Download(fn func(r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := storage.FromContext(r.Context())
		if !ok {
			http.Error(w, "missing storage in context", http.StatusInternalServerError)
			return
		}
		if err := s.ServeFile(w, r, fn(r)); err != nil {
			storage.Error(w, err)
			return
		}
	}
}

func Delete(fn func(t *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := storage.FromContext(ctx)
		if !ok {
			http.Error(w, "missing storage in context", http.StatusInternalServerError)
			return
		}
		if err := s.Delete(ctx, fn(r)); err != nil {
			storage.Error(w, err)
			return
		}
	}
}
