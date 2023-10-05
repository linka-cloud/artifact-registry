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

package deb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	"go.linka.cloud/artifact-registry/pkg/logger"
	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

var _ packages.Provider = (*provider)(nil)

func init() {
	packages.Register("deb", newProvider)
}

func newProvider(_ context.Context) (packages.Provider, error) {
	return &provider{}, nil
}

type provider struct{}

func (p *provider) Repository() storage.Repository {
	return &repo{}
}

func (p *provider) setup(w http.ResponseWriter, r *http.Request) {
	repo, dist, component := mux.Vars(r)["repo"], mux.Vars(r)["distribution"], mux.Vars(r)["component"]
	user, pass, _ := r.BasicAuth()
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	args := setupArgs{
		Name:      strings.Replace(repo, "/", "-", -1),
		User:      user,
		Password:  pass,
		Scheme:    scheme,
		Host:      r.Host,
		Path:      strings.TrimSuffix(r.URL.Path, fmt.Sprintf("/%s/%s/setup", dist, component)),
		Dist:      dist,
		Component: component,
	}
	if err := scriptTemplate.Execute(w, args); err != nil {
		logger.C(r.Context()).WithError(err).Error("failed to execute template")
	}
}

func (p *provider) Register(m *mux.Router) {
	m.HandleFunc("/{repo:.+}/{distribution}/{component}/setup", p.setup).Methods(http.MethodGet)
	m.HandleFunc("/{repo:.+}/"+RepositoryPublicKey, packages.Download(func(r *http.Request) string {
		return RepositoryPublicKey
	})).Methods(http.MethodGet)
	m.HandleFunc("/{repo:.+}/dists/{distribution}/{filename}", packages.Download(func(r *http.Request) string {
		dist, filename := mux.Vars(r)["distribution"], mux.Vars(r)["filename"]
		return filepath.Join("dists", dist, filename)
	})).Methods(http.MethodGet)
	m.HandleFunc("/{repo:.+}/dists/{distribution}/{component}/{architecture}/{filename}", packages.Download(func(r *http.Request) string {
		dist, component, architecture, filename := mux.Vars(r)["distribution"], mux.Vars(r)["component"], mux.Vars(r)["architecture"], mux.Vars(r)["filename"]
		return filepath.Join("dists", dist, component, architecture, filename)
	})).Methods(http.MethodGet)

	m.HandleFunc("/{repo:.+}/pool/{distribution}/{component}/upload", packages.Upload(func(r *http.Request, reader io.Reader, size int64, key string) (storage.Artifact, error) {
		distribution, component := mux.Vars(r)["distribution"], mux.Vars(r)["component"]
		return NewPackage(reader, distribution, component, size)
	})).Methods(http.MethodPut)
	m.HandleFunc("/{repo:.+}/pool/{distribution}/{component}/{name}_{version}_{architecture}.deb", packages.Download(func(r *http.Request) string {
		dist, component, name, version, architecture := mux.Vars(r)["distribution"], mux.Vars(r)["component"], mux.Vars(r)["name"], mux.Vars(r)["version"], mux.Vars(r)["architecture"]
		return filepath.Join("pool", dist, component, name+"_"+version+"_"+architecture+".deb")
	})).Methods(http.MethodGet)
	m.HandleFunc("/{repo:.+}/pool/{distribution}/{component}/{name}/{version}/{architecture}", packages.Delete(func(r *http.Request) string {
		dist, component, name, version, architecture := mux.Vars(r)["distribution"], mux.Vars(r)["component"], mux.Vars(r)["name"], mux.Vars(r)["version"], mux.Vars(r)["architecture"]
		return filepath.Join("pool", dist, component, name+"_"+version+"_"+architecture+".deb")
	})).Methods(http.MethodDelete)

}
