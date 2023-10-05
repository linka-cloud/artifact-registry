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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"go.linka.cloud/artifact-registry/pkg/logger"
	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/storage"
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

func newProvider(_ context.Context) (packages.Provider, error) {
	return &provider{}, nil
}

type provider struct{}

func (p *provider) Register(r *mux.Router) {
	r.HandleFunc("/{repo:.+}.repo", p.config).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/setup", p.setup).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/upload", p.upload()).Methods(http.MethodPut)
	r.HandleFunc("/{repo:.+}/repodata/{filename}", p.download()).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/{filename}", p.download()).Methods(http.MethodGet)
	r.HandleFunc("/{repo:.+}/{filename}", p.delete()).Methods(http.MethodDelete)
}

func (p *provider) Repository() storage.Repository {
	return &repo{}
}

func (p *provider) config(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := mux.Vars(r)["repo"]
	if _, ok := storage.FromContext(ctx); !ok {
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

func (p *provider) setup(w http.ResponseWriter, r *http.Request) {
	repo := mux.Vars(r)["repo"]
	user, pass, _ := r.BasicAuth()
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	args := setupArgs{
		Name:     strings.Replace(repo, "/", "-", -1),
		User:     user,
		Password: pass,
		Scheme:   scheme,
		Host:     r.Host,
		Path:     strings.TrimSuffix(r.URL.Path, "/setup"),
	}
	if err := scriptTemplate.Execute(w, args); err != nil {
		logger.C(r.Context()).WithError(err).Error("failed to execute template")
	}
}

func (p *provider) upload() http.HandlerFunc {
	return packages.Upload(func(r *http.Request, reader io.Reader, size int64, key string) (storage.Artifact, error) {
		return NewPackage(reader, size, key)
	})
}

func (p *provider) download() http.HandlerFunc {
	return packages.Download(func(r *http.Request) string {
		return mux.Vars(r)["filename"]
	})
}

func (p *provider) delete() http.HandlerFunc {
	return packages.Delete(func(r *http.Request) string {
		return mux.Vars(r)["filename"]
	})
}
