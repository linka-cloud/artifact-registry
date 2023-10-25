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

package storage

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.linka.cloud/grpc-toolkit/logger"

	"go.linka.cloud/artifact-registry/pkg/auth"
)

type MiddlewareFunc = func(repoVar string) mux.MiddlewareFunc

func Middleware(ar Repository) MiddlewareFunc {
	return func(repoVar string) mux.MiddlewareFunc {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := auth.Context(r.Context(), r)
				name := mux.Vars(r)[repoVar]
				if name == "" {
					n := Options(ctx).repo
					if n == "" {
						http.Error(w, "missing repository name", http.StatusBadRequest)
						return
					}
					name = n
				}
				ctx = logger.Set(ctx, logger.C(ctx).WithField("repo", name).WithField("type", ar.Name()))
				s, err := NewStorage(ctx, name, ar)
				if err != nil {
					Error(w, err)
					return
				}
				defer s.Close()
				next.ServeHTTP(w, r.WithContext(Context(ctx, s)))
			})
		}
	}
}
