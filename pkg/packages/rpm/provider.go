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

func (p *provider) Repository() storage.Repository {
	return &repo{}
}

func (p *provider) config(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := mux.Vars(r)["repo"]
	if _, err := storage.FromContext(ctx).Stat(ctx, RepositoryPublicKey); err != nil {
		storage.Error(w, err)
		return
	}
	host := strings.TrimSuffix(r.Host, "/")
	if u, p, ok := r.BasicAuth(); ok {
		host = fmt.Sprintf("%s:%s@%s", u, p, host)
	}
	url := fmt.Sprintf("%s://%s/%s", packages.Scheme(r), host, strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, ".repo"), "/"))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(definition, strings.NewReplacer("/", "-").Replace(name), url, RepositoryPublicKey)))
}

func (p *provider) setup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, err := storage.FromContext(ctx).Stat(ctx, RepositoryPublicKey); err != nil {
		storage.Error(w, err)
		return
	}
	repo := mux.Vars(r)["repo"]
	user, pass, _ := r.BasicAuth()
	args := SetupArgs{
		Name:     strings.Replace(repo, "/", "-", -1),
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

func (p *provider) Routes() []*packages.Route {
	return []*packages.Route{
		{
			Method:  http.MethodGet,
			Path:    "/{repo:.+}.repo",
			Handler: p.config,
		},
		{
			Method:  http.MethodGet,
			Path:    "/{repo:.+}/setup",
			Handler: p.setup,
		},
		{
			Method: http.MethodPut,
			Path:   "/{repo:.+}/push",
			Handler: packages.Push(func(r *http.Request, reader io.Reader, size int64, key string) (storage.Artifact, error) {
				return NewPackage(reader, size, key)
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/{repo:.+}/repodata/{filename}",
			Handler: packages.Pull(func(r *http.Request) string {
				return mux.Vars(r)["filename"]
			}),
		},
		{
			Method: http.MethodGet,
			Path:   "/{repo:.+}/{filename}",
			Handler: packages.Pull(func(r *http.Request) string {
				return mux.Vars(r)["filename"]
			}),
		},
		{
			Method: http.MethodDelete,
			Path:   "/{repo:.+}/{filename}",
			Handler: packages.Delete(func(r *http.Request) string {
				return mux.Vars(r)["filename"]
			}),
		},
	}
}
