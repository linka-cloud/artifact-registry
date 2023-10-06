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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2/registry/remote"

	cache2 "go.linka.cloud/artifact-registry/pkg/cache"
	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/slices"
	"go.linka.cloud/artifact-registry/pkg/storage"
	"go.linka.cloud/artifact-registry/pkg/storage/auth"
)

var cache = cache2.New()

type Stats struct {
	Size  int64 `json:"size"`
	Count int64 `json:"count"`
}

type Repository struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Size        int64      `json:"size"`
	LastUpdated *time.Time `json:"lastUpdated"`
	Metadata    Stats      `json:"metadata"`
	Packages    Stats      `json:"packages"`
}

func Auth(w http.ResponseWriter, r *http.Request) {
	ctx := auth.Context(r.Context(), r)
	o := storage.Options(ctx)
	name, typ := mux.Vars(r)["repo"], mux.Vars(r)["type"]
	reg, err := remote.NewRegistry(o.Host())
	if err != nil {
		storage.Error(w, err)
		return
	}
	reg.Client = o.Client(ctx, o.Host())
	repo, err := reg.Repository(ctx, name)
	if err != nil {
		storage.Error(w, err)
		return
	}
	var ts []string
	if typ == "" {
		ts = packages.Providers()
	} else {
		ts = []string{typ}
	}
	for _, v := range ts {
		if _, err := repo.Manifests().Resolve(ctx, v); err != nil {
			if storage.IsNotFound(err) {
				continue
			}
			if err != nil {
				storage.Error(w, err)
				return
			}
		}
		return
	}
	http.Error(w, "No repository found", http.StatusNotFound)
}

func listImageRepositories(ctx context.Context, reg *remote.Registry, name string, typ ...string) ([]*Repository, error) {
	repo, err := reg.Repository(ctx, name)
	if err != nil {
		if storage.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	typ = slices.Filter(typ, func(s string) bool {
		return s != ""
	})
	var tags []string
	if len(typ) == 0 {
		tags = packages.Providers()
	} else {
		tags = typ
	}
	var (
		out []*Repository
		mu  sync.Mutex
	)
	g, ctx := errgroup.WithContext(ctx)
	fn := func(i int, typ string) error {
		desc, err := repo.Manifests().Resolve(ctx, typ)
		if err != nil {
			if storage.IsNotFound(err) {
				return nil
			}
			return err
		}
		var m ocispec.Manifest
		v, ok := cache.Get(desc.Digest.String())
		if !ok {
			rc, err := repo.Manifests().Fetch(ctx, desc)
			if err != nil {
				return err
			}
			defer rc.Close()
			if err := json.NewDecoder(rc).Decode(&m); err != nil {
				return err
			}
		} else {
			m = v.(ocispec.Manifest)
		}
		cache.Set(desc.Digest.String(), m, cache2.WithTTL(5*time.Minute))
		t, err := time.Parse(time.RFC3339, m.Annotations[ocispec.AnnotationCreated])
		if err != nil {
			return err
		}
		r := &Repository{
			Name:        storage.Options(ctx).Host() + "/" + name,
			Type:        typ,
			LastUpdated: &t,
		}
		for _, v := range m.Layers {
			r.Size += v.Size
			if v.MediaType == "application/vnd.lk.registry.layer.v1."+typ {
				r.Packages.Size += v.Size
				r.Packages.Count++
			} else {
				r.Metadata.Size += v.Size
				r.Metadata.Count++
			}
		}
		mu.Lock()
		out = append(out, r)
		mu.Unlock()
		return nil
	}
	for i, v := range tags {
		i, v := i, v
		g.Go(func() error {
			if err := fn(i, v); err != nil {
				return fmt.Errorf("%s: %w", v, err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func ListRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := auth.Context(r.Context(), r)
	o := storage.Options(ctx)
	typ := mux.Vars(r)["type"]
	reg, err := remote.NewRegistry(o.Host())
	if err != nil {
		storage.Error(w, err)
		return
	}
	reg.Client = o.Client(ctx, o.Host())
	var repos []string
	if err := reg.Repositories(ctx, "", func(r []string) error {
		repos = append(repos, r...)
		return nil
	}); err != nil {
		storage.Error(w, err)
		return
	}
	var out []*Repository
	var mu sync.Mutex
	g, ctx := errgroup.WithContext(ctx)
	for _, v := range repos {
		v := v
		g.Go(func() error {
			rs, err := listImageRepositories(ctx, reg, v, typ)
			if err != nil {
				return fmt.Errorf("%s: %w", v, err)
			}
			mu.Lock()
			out = append(out, rs...)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		storage.Error(w, err)
		return
	}
	if err := json.NewEncoder(w).Encode(out); err != nil {
		storage.Error(w, err)
		return
	}
}

func ListImageRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := auth.Context(r.Context(), r)
	o := storage.Options(ctx)
	name, typ := mux.Vars(r)["repo"], mux.Vars(r)["type"]
	reg, err := remote.NewRegistry(o.Host())
	if err != nil {
		storage.Error(w, err)
		return
	}
	reg.Client = o.Client(ctx, o.Host())
	out, err := listImageRepositories(ctx, reg, name, typ)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		storage.Error(w, err)
		return
	}
}

func Packages(w http.ResponseWriter, r *http.Request) {
	ctx := auth.Context(r.Context(), r)
	typ, repo := mux.Vars(r)["type"], mux.Vars(r)["repo"]
	p, err := packages.New(ctx, typ)
	if err != nil {
		storage.Error(w, err)
		return
	}
	s, err := storage.NewStorage(ctx, repo, p.Repository())
	if err != nil {
		storage.Error(w, err)
		return
	}
	pkgs, err := s.Artifacts(ctx)
	if err != nil {
		storage.Error(w, err)
		return
	}
	if err := json.NewEncoder(w).Encode(pkgs); err != nil {
		storage.Error(w, err)
		return
	}
}

func Init(_ context.Context, r *mux.Router, domain string) error {
	r.Path("/_auth/{repo:.+}").Methods(http.MethodGet, http.MethodPost).HandlerFunc(Auth)
	r.Path("/_repositories").Methods(http.MethodGet).HandlerFunc(ListRepositories)
	r.Path("/_repositories/{repo:.+}").Methods(http.MethodGet).HandlerFunc(ListImageRepositories)
	subs := []*mux.Router{r.PathPrefix("/{type}/_packages/").Subrouter()}
	if domain != "" {
		subs = append(subs, r.Host("{type}."+domain+"/_packages").Subrouter())
	}
	for _, v := range subs {
		v.Path("/{repo:.+}").Methods(http.MethodGet).HandlerFunc(Packages)
	}
	return nil
}
