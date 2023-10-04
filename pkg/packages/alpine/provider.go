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

package alpine

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"

	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

var _ packages.Provider = (*provider)(nil)

func init() {
	packages.Register("apk", newProvider)
}

func newProvider(_ context.Context) (packages.Provider, error) {
	return &provider{}, nil
}

type provider struct{}

func (p *provider) Repository() storage.Repository {
	return &repo{}
}

func (p *provider) DownloadKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s, ok := storage.FromContext(ctx)
	if !ok {
		http.Error(w, "missing storage in context", http.StatusInternalServerError)
		return
	}
	pub, f, err := PublicKeyAndFingerprintFromPrivateKey(s.Key())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pub)))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%s`, fmt.Sprintf("lkar@%s.rsa.pub", hex.EncodeToString(f))))
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	io.Copy(w, bytes.NewReader(pub))
}

func (p *provider) Register(m *mux.Router) {
	m.HandleFunc("/{repo:.+}/{branch}/{repository}/key", p.DownloadKey).Methods(http.MethodGet)
	m.HandleFunc("/{repo:.+}/{branch}/{repository}/upload", packages.Upload(func(r *http.Request, reader io.Reader, size int64, key string) (storage.Artifact, error) {
		branch, repo := mux.Vars(r)["branch"], mux.Vars(r)["repository"]
		return NewPackage(reader, branch, repo, size)
	})).Methods(http.MethodPut)
	m.HandleFunc("/{repo:.+}/{branch}/{repository}/{architecture}/{filename}", packages.Download(func(r *http.Request) string {
		branch, repo, arch, filename := mux.Vars(r)["branch"], mux.Vars(r)["repository"], mux.Vars(r)["architecture"], mux.Vars(r)["filename"]
		return filepath.Join(branch, repo, arch, filename)
	})).Methods(http.MethodGet)
	m.HandleFunc("/{repo:.+}/{branch}/{repository}/{architecture}/{filename}", packages.Delete(func(r *http.Request) string {
		branch, repo, arch, filename := mux.Vars(r)["branch"], mux.Vars(r)["repository"], mux.Vars(r)["architecture"], mux.Vars(r)["filename"]
		return filepath.Join(branch, repo, arch, filename)
	})).Methods(http.MethodDelete)
}
