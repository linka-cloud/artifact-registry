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
	"go.linka.cloud/grpc-toolkit/logger"

	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

const Name = "deb"

var _ packages.Provider = (*provider)(nil)

func init() {
	packages.Register(Name, newProvider)
}

func newProvider(_ context.Context) (packages.Provider, error) {
	return &provider{}, nil
}

type provider struct{}

func (p *provider) Repository() storage.Repository {
	return &repo{}
}

func (p *provider) setup(_ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s := storage.FromContext(ctx)
		if _, err := s.Stat(ctx, RepositoryPublicKey); err != nil {
			storage.Error(w, err)
			return
		}
		repo, dist, component := mux.Vars(r)["repo"], mux.Vars(r)["distribution"], mux.Vars(r)["component"]
		var name string
		if repo != "" {
			name = strings.NewReplacer("/", "-").Replace(repo)
		} else {
			name = strings.NewReplacer("/", "-", ".", "-").Replace(strings.TrimPrefix(strings.Split(r.Host, ":")[0], Name+"."))
		}
		user, pass, _ := r.BasicAuth()
		args := SetupArgs{
			Name:      name,
			User:      user,
			Password:  pass,
			Scheme:    packages.Scheme(r),
			Host:      r.Host,
			Path:      strings.TrimSuffix(r.URL.Path, fmt.Sprintf("/%s/%s/setup", dist, component)),
			Dist:      dist,
			Component: component,
		}
		if err := scriptTemplate.Execute(w, args); err != nil {
			logger.C(r.Context()).WithError(err).Error("failed to execute template")
		}
	}
}

func (p *provider) Routes() []*packages.Route {
	return []*packages.Route{
		{
			Method:  http.MethodGet,
			Path:    "/{distribution}/{component}/setup",
			Handler: p.setup,
		},
		{
			Method: http.MethodGet,
			Path:   "/" + RepositoryPublicKey,
			Handler: packages.Pull(func(r *http.Request) string {
				return RepositoryPublicKey
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/dists/{distribution}/{filename}",
			Handler: packages.Pull(func(r *http.Request) string {
				dist, filename := mux.Vars(r)["distribution"], mux.Vars(r)["filename"]
				return filepath.Join("dists", dist, filename)
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/dists/{distribution}/{component}/{architecture}/{filename}",
			Handler: packages.Pull(func(r *http.Request) string {
				dist, component, architecture, filename := mux.Vars(r)["distribution"], mux.Vars(r)["component"], mux.Vars(r)["architecture"], mux.Vars(r)["filename"]
				return filepath.Join("dists", dist, component, architecture, filename)
			}),
		},
		{
			Method: http.MethodPut,
			Path:   "/pool/{distribution}/{component}/push",
			Handler: packages.Push(func(r *http.Request, reader io.Reader, size int64, key string) (storage.Artifact, error) {
				distribution, component := mux.Vars(r)["distribution"], mux.Vars(r)["component"]
				return NewPackage(reader, distribution, component, size)
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/pool/{distribution}/{component}/{name}_{version}_{architecture}.deb",
			Handler: packages.Pull(func(r *http.Request) string {
				dist, component, name, version, architecture := mux.Vars(r)["distribution"], mux.Vars(r)["component"], mux.Vars(r)["name"], mux.Vars(r)["version"], mux.Vars(r)["architecture"]
				return filepath.Join("pool", dist, component, name+"_"+version+"_"+architecture+".deb")
			}),
		},
		{
			Method: http.MethodDelete,
			Path:   "/pool/{distribution}/{component}/{name}_{version}_{architecture}.deb",
			Handler: packages.Delete(func(r *http.Request) string {
				dist, component, name, version, architecture := mux.Vars(r)["distribution"], mux.Vars(r)["component"], mux.Vars(r)["name"], mux.Vars(r)["version"], mux.Vars(r)["architecture"]
				return filepath.Join("pool", dist, component, name+"_"+version+"_"+architecture+".deb")
			}),
		},
	}
}
