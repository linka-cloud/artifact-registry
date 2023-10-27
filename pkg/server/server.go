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

package server

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/grpc-toolkit/react"

	"go.linka.cloud/artifact-registry/pkg/api"
	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/storage"
	"go.linka.cloud/artifact-registry/ui"
)

func Run(ctx context.Context, addr, aesKey, backend, domain, repo, cert, key string, disableUI bool, opts ...storage.Option) error {
	if aesKey == "" {
		return fmt.Errorf("aesKey is required")
	}
	logger.C(ctx).Infof("initializing artifact registry using backend %s", backend)
	router := mux.NewRouter().StrictSlash(true)
	k := sha256.Sum256([]byte(aesKey))
	ctx = storage.WithOptions(ctx, append(opts, storage.WithKey(k[:]))...)

	if !disableUI {
		uih, err := react.NewHandler(ui.UI, "build")
		if err != nil {
			return err
		}
		router.PathPrefix("/ui").Handler(http.StripPrefix("/ui", uih))
		router.Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/ui", http.StatusFound)
		})
	}

	router.Path("/_/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if err := api.Init(ctx, router, domain, repo); err != nil {
		return err
	}
	if err := packages.Init(ctx, router, domain, repo); err != nil {
		return err
	}

	if err := router.Walk(func(r *mux.Route, _ *mux.Router, _ []*mux.Route) error { return r.GetError() }); err != nil {
		return err
	}

	s := http.Server{
		BaseContext: func(lis net.Listener) context.Context {
			return ctx
		},
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrap := wrap(w)
			start := time.Now()
			remote := strings.Split(r.RemoteAddr, ":")[0]
			for _, v := range r.Header["X-Forwarded-For"] {
				if ip := net.ParseIP(v); ip != nil && !ip.IsPrivate() {
					remote = v
					break
				}
			}
			log := logger.C(r.Context()).WithFields(
				"method", r.Method,
				"path", r.URL.Path,
				"remote", remote,
				"userAgent", r.UserAgent(),
			)
			if u, _, ok := r.BasicAuth(); ok {
				log = log.WithField("user", u)
			}
			time.Since(start)
			router.ServeHTTP(wrap, r.WithContext(logger.Set(r.Context(), log)))
			log = log.WithFields(
				"duration", time.Since(start),
				"status", http.StatusText(wrap.status),
				"statusCode", wrap.status,
				"responseSize", wrap.size,
			)
			if wrap.status == 0 {
				wrap.status = 200
			}
			if wrap.status < 400 {
				log.Info("")
			} else {
				log.Error(wrap.body.String())
			}
		}),
	}
	logger.C(ctx).Infof("starting server at %s", addr)
	if cert != "" && key != "" {
		return s.ListenAndServeTLS(cert, key)
	}
	return s.ListenAndServe()
}
