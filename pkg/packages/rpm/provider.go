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

package rpm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/repository"
)

const definition = `[%[1]s]
name=%[1]s
baseurl=%[2]s
enabled=1
gpgcheck=1
gpgkey=%[2]s/%[3]s
`

var _ packages.Provider = (*provider)(nil)

func init() {
	packages.Register("rpm", newProvider)
}

func newProvider(_ context.Context, backend string, key []byte) (packages.Provider, error) {
	return &provider{mdwl: repository.StorageMiddleware[*Package, *repo](&repo{}, backend, key)}, nil
}

type provider struct {
	mdwl repository.StorageMiddlewareFunc
}

func (p *provider) Register(r *mux.Router) {
	r.Use(p.mdwl("repo"))
	r.HandleFunc("/{repo:.+}.repo", p.repositoryConfig).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/"+RepositoryPublicKey, p.repositoryKey).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/upload", p.uploadPackage).Methods(http.MethodPut)
	r.HandleFunc("/{repo:.+}/repodata/{filename}", p.repositoryFile).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/{filename}", p.downloadPackage).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/{filename}", p.deletePackage).Methods(http.MethodDelete)
}

func (p *provider) repositoryConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := mux.Vars(r)["repo"]
	if _, ok := repository.FromContext[*Package, *repo](ctx); !ok {
		http.Error(w, "missing storage in context", http.StatusInternalServerError)
		return
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := strings.TrimSuffix(r.Host, "/")
	if u, p, ok := r.BasicAuth(); ok {
		host = fmt.Sprintf("%s:%s@%s", u, p, host)
	}
	url := fmt.Sprintf("%s://%s/%s", scheme, host, strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, ".repo"), "/"))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(definition, strings.NewReplacer("/", "-").Replace(name), url, RepositoryPublicKey)))
}

func (p *provider) repositoryKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repo, ok := repository.FromContext[*Package, *repo](ctx)
	if !ok {
		http.Error(w, "missing storage in context", http.StatusInternalServerError)
		return
	}
	rc, err := repo.Open(ctx, RepositoryPublicKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.Copy(w, rc)
}

func (p *provider) uploadPackage(w http.ResponseWriter, r *http.Request) {
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
	repo, ok := repository.FromContext[*Package, *repo](ctx)
	if !ok {
		http.Error(w, "missing storage in context", http.StatusInternalServerError)
		return
	}
	pkg, err := NewPackage(reader, size, repo.Key())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := repo.Write(ctx, pkg); err != nil {
		if errors.Is(err, os.ErrExist) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (p *provider) downloadPackage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	file := mux.Vars(r)["filename"]
	repo, ok := repository.FromContext[*Package, *repo](ctx)
	if !ok {
		http.Error(w, "missing storage in context", http.StatusInternalServerError)
		return
	}
	if err := repo.ServeFile(w, r, file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *provider) deletePackage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	file := mux.Vars(r)["filename"]
	repo, ok := repository.FromContext[*Package, *repo](ctx)
	if !ok {
		http.Error(w, "missing storage in context", http.StatusInternalServerError)
		return
	}
	if err := repo.Delete(ctx, file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *provider) repositoryFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	file := mux.Vars(r)["filename"]
	repo, ok := repository.FromContext[*Package, *repo](ctx)
	if !ok {
		http.Error(w, "missing storage in context", http.StatusInternalServerError)
		return
	}
	if err := repo.ServeFile(w, r, file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
