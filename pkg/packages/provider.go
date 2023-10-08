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
	"errors"
	"fmt"

	"github.com/gorilla/mux"

	"go.linka.cloud/artifact-registry/pkg/storage"
)

var ErrUnknownProvider = errors.New("unknown provider")

type Provider interface {
	Routes() []*Route
	Repository() storage.Repository
}

type ProviderFactory func(ctx context.Context) (Provider, error)

var providers = map[string]ProviderFactory{}

func Register(name string, factory ProviderFactory) {
	providers[name] = factory
}

func Providers() []string {
	var ret []string
	for k := range providers {
		ret = append(ret, k)
	}
	return ret
}

func New(ctx context.Context, name string) (Provider, error) {
	f, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("%s: %w", name, ErrUnknownProvider)
	}
	return f(ctx)
}

func Init(ctx context.Context, r *mux.Router, domain string) error {
	for k, v := range providers {
		p, err := v(ctx)
		if err != nil {
			return err
		}
		mdlw := storage.Middleware(p.Repository())("repo")
		subs := []*mux.Router{r.PathPrefix("/" + k).Subrouter()}
		if domain != "" {
			subs = append(subs, r.Host(k+"."+domain).Subrouter())
		}
		for _, v := range subs {
			v.Use(mdlw)
			for _, vv := range p.Routes() {
				if err := v.Path(vv.Path).Methods(vv.Method).HandlerFunc(vv.Handler).GetError(); err != nil {
					return fmt.Errorf("%s: %q: %w", k, vv.Path, err)
				}
			}
		}
	}
	return nil
}
