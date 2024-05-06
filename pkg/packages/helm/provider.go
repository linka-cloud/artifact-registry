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

package helm

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.linka.cloud/grpc-toolkit/logger"

	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

const Name = "helm"

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
		if _, err := storage.FromContext(ctx).Stat(ctx, RepositoryPublicKey); err != nil {
			storage.Error(w, err)
			return
		}
		user, pass, _ := r.BasicAuth()
		repo := mux.Vars(r)["repo"]
		var name string
		if repo != "" {
			name = strings.NewReplacer("/", "-").Replace(repo)
		} else {
			name = strings.NewReplacer("/", "-", ".", "-").Replace(strings.TrimPrefix(strings.Split(r.Host, ":")[0], Name+"."))
		}
		args := SetupArgs{
			Name:     name,
			User:     user,
			Password: pass,
			Scheme:   packages.Scheme(r),
			Host:     r.Host,
			Path:     strings.TrimSuffix(r.URL.Path, "/setup"),
		}
		if err := scriptTemplate.Execute(w, args); err != nil {
			logger.C(r.Context()).WithError(err).Error("failed to execute template")
		}
	}
}

func (p *provider) Routes() []*packages.Route {
	return []*packages.Route{
		{
			Path:    "/setup",
			Method:  http.MethodGet,
			Handler: p.setup,
		},
		{
			Path:   "/{filename}",
			Method: http.MethodGet,
			Handler: packages.Pull(func(r *http.Request) string {
				return mux.Vars(r)["filename"]
			}),
		},
		{
			Path:   "/push",
			Method: http.MethodPut,
			Handler: packages.Push(func(r *http.Request, reader io.Reader, key string) (storage.Artifact, error) {
				return NewPackage(reader)
			}),
		},
		{
			Path:   "/{filename}",
			Method: http.MethodDelete,
			Handler: packages.Delete(func(r *http.Request) string {
				return mux.Vars(r)["filename"]
			}),
		},
	}
}
