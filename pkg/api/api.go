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

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.linka.cloud/grpc-toolkit/logger"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2/registry"

	"go.linka.cloud/artifact-registry/pkg/auth"
	"go.linka.cloud/artifact-registry/pkg/cache"
	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/slices"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

const sessionName = "auth"

type Stats struct {
	Size  int64 `json:"size"`
	Count int64 `json:"count"`
}

type Repository struct {
	Name        string     `json:"name,omitempty"`
	Type        string     `json:"type"`
	Size        int64      `json:"size"`
	LastUpdated *time.Time `json:"lastUpdated"`
	Metadata    Stats      `json:"metadata"`
	Packages    Stats      `json:"packages"`
}

type handler struct {
	store sessions.Store
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := auth.Context(r.Context(), r)
	name, typ := mux.Vars(r)["repo"], mux.Vars(r)["type"]
	o := storage.Options(ctx)
	if n := o.Repo(); name == "" && n != "" {
		name = n
	}
	reg, err := o.NewRegistry(ctx)
	if err != nil {
		storage.Error(w, err)
		return
	}
	if name == "" {
		skip := errors.New("skip")
		if err := reg.Repositories(ctx, "", func(r []string) error {
			return skip
		}); err != nil && !errors.Is(err, skip) {
			storage.Error(w, err)
			return
		}
		h.saveCredentials(w, r)
		return
	}
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
		h.saveCredentials(w, r)
		return
	}
	h.saveCredentials(w, r)
}

func (h *handler) saveCredentials(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return
	}
	s, err := h.store.Get(r, sessionName)
	if err != nil {
		logger.C(r.Context()).WithError(err).Error("failed to get session")
		s.Options.MaxAge = -1
		if err := s.Save(r, w); err != nil {
			logger.C(r.Context()).WithError(err).Error("failed to delete session")
		}
		return
	}
	s.Values["user"] = user
	s.Values["pass"] = pass
	if err := s.Save(r, w); err != nil {
		logger.C(r.Context()).WithError(err).Error("failed to delete session")
	}
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	s, err := h.store.Get(r, sessionName)
	if err != nil {
		return
	}
	s.Options.MaxAge = -1
	if err := s.Save(r, w); err != nil {
		logger.C(r.Context()).WithError(err).Error("failed to save session")
	}
}

type credentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func (h *handler) Credentials(w http.ResponseWriter, r *http.Request) {
	u, p, _ := r.BasicAuth()
	if err := json.NewEncoder(w).Encode(credentials{User: u, Password: p}); err != nil {
		storage.Error(w, err)
		return
	}
}

func listImageRepositories(ctx context.Context, reg registry.Registry, name string, typ ...string) ([]*Repository, error) {
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
		cache.Set(desc.Digest.String(), m, cache.WithTTL(cache.DefaultTTL))
		t, err := time.Parse(time.RFC3339, m.Annotations[ocispec.AnnotationCreated])
		if err != nil {
			return err
		}
		// do not leak the repo name in single repo mode
		var repo string
		if opts := storage.Options(ctx); opts.Repo() == "" {
			repo = opts.Host() + "/" + name
		}
		r := &Repository{
			Name:        repo,
			Type:        typ,
			LastUpdated: &t,
		}
		l := make(map[string]struct{})
		for _, v := range m.Layers {
			_, seen := l[v.Digest.String()]
			l[v.Digest.String()] = struct{}{}
			if !seen {
				r.Size += v.Size
			}
			if v.MediaType == "application/vnd.lk.registry.layer.v1."+typ {
				if !seen {
					r.Packages.Size += v.Size
				}
				r.Packages.Count++
			} else {
				if !seen {
					r.Metadata.Size += v.Size
				}
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

func (h *handler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := auth.Context(r.Context(), r)
	typ := mux.Vars(r)["type"]
	reg, err := storage.Options(ctx).NewRegistry(ctx)
	if err != nil {
		storage.Error(w, err)
		return
	}
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
	if len(out) == 0 {
		out = []*Repository{}
	}
	if err := json.NewEncoder(w).Encode(out); err != nil {
		storage.Error(w, err)
		return
	}
}

func (h *handler) ListImageRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := auth.Context(r.Context(), r)
	name, typ := mux.Vars(r)["repo"], mux.Vars(r)["type"]
	o := storage.Options(ctx)
	if n := o.Repo(); name == "" && n != "" {
		name = n
	}
	reg, err := o.NewRegistry(ctx)
	if err != nil {
		storage.Error(w, err)
		return
	}
	out, err := listImageRepositories(ctx, reg, name, typ)
	if err != nil {
		storage.Error(w, err)
		return
	}
	if len(out) == 0 {
		out = []*Repository{}
	}
	if err := json.NewEncoder(w).Encode(out); err != nil {
		storage.Error(w, err)
		return
	}
}

func (h *handler) Packages(w http.ResponseWriter, r *http.Request) {
	ctx := auth.Context(r.Context(), r)
	repo, typ := mux.Vars(r)["repo"], mux.Vars(r)["type"]
	if n := storage.Options(ctx).Repo(); repo == "" && n != "" {
		repo = n
	}
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

func (h *handler) cookie2Basic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if we already have basic auth, we can skip
		if _, _, ok := r.BasicAuth(); ok {
			next.ServeHTTP(w, r)
			return
		}
		s, err := h.store.Get(r, "auth")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		u, ok1 := s.Values["user"]
		p, ok2 := s.Values["pass"]
		if ok1 && ok2 {
			r.SetBasicAuth(u.(string), p.(string))
		}
		next.ServeHTTP(w, r)
	})
}

func Init(ctx context.Context, r *mux.Router, domain, repo string) error {
	tregx := strings.Join(packages.Providers(), "|")
	h := &handler{store: sessions.NewCookieStore(storage.Options(ctx).Key())}
	r.Use(h.cookie2Basic)
	r.Path("/_auth/login").Methods(http.MethodGet, http.MethodPost).HandlerFunc(h.Login)
	if repo == "" {
		r.Path("/_auth/{repo:.+}/login").Methods(http.MethodGet, http.MethodPost).HandlerFunc(h.Login)
	}
	r.Path("/_auth/logout").Methods(http.MethodGet, http.MethodPost).HandlerFunc(h.Logout)
	// TODO(adphi): we should find a way to protect this to make it only available to the browser
	r.Path("/_auth/credentials").Methods(http.MethodGet).HandlerFunc(h.Credentials)

	if repo == "" {
		r.Host(fmt.Sprintf("{type:%s}.%s", tregx, domain)).Path("/_repositories").Methods(http.MethodGet).HandlerFunc(h.ListRepositories)
		r.Host(fmt.Sprintf("{type:%s}.%s", tregx, domain)).Path("/_repositories/{repo:.+}").Methods(http.MethodGet).HandlerFunc(h.ListImageRepositories)
		r.Path("/_repositories/{repo:.+}").Methods(http.MethodGet).HandlerFunc(h.ListImageRepositories)
		r.Path("/_repositories").Methods(http.MethodGet).HandlerFunc(h.ListRepositories)

		r.Host(fmt.Sprintf("{type:%s}.%s", tregx, domain)).Path("/_packages/{repo:.+}").Methods(http.MethodGet).HandlerFunc(h.Packages)
		r.Path(fmt.Sprintf("/_packages/{type:%s}/{repo:.+}", tregx)).Methods(http.MethodGet).HandlerFunc(h.Packages)
	} else {
		r.Host(fmt.Sprintf("{type:%s}.%s", tregx, domain)).Path("/_repositories").Methods(http.MethodGet).HandlerFunc(h.ListImageRepositories)
		r.Path("/_repositories").Methods(http.MethodGet).HandlerFunc(h.ListImageRepositories)

		r.Host(fmt.Sprintf("{type:%s}.%s", tregx, domain)).Path("/_packages").Methods(http.MethodGet).HandlerFunc(h.Packages)
		r.Path(fmt.Sprintf("/_packages/{type:%s}", tregx)).Methods(http.MethodGet).HandlerFunc(h.Packages)
	}
	return nil
}
