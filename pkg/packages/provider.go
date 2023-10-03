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

package packages

import (
	"context"

	"github.com/gorilla/mux"

	"go.linka.cloud/artifact-registry/pkg/storage"
)

type Provider interface {
	Register(m *mux.Router)
	Repository() storage.Repository
}

type ProviderFactory func(ctx context.Context) (Provider, error)

var providers = map[string]ProviderFactory{}

func Register(name string, factory ProviderFactory) {
	providers[name] = factory
}

func Init(ctx context.Context, r *mux.Router, backend string, key []byte, domain string) error {
	for k, v := range providers {
		p, err := v(ctx)
		if err != nil {
			return err
		}
		subs := []*mux.Router{r.PathPrefix("/" + k).Subrouter()}
		if domain != "" {
			subs = append(subs, r.Host(k+"."+domain).Subrouter())
		}
		for _, v := range subs {
			v.Use(storage.Middleware(p.Repository(), backend, key)("repo"))
			p.Register(v)
		}
	}
	return nil
}
