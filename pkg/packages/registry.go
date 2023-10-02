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
)

type Provider interface {
	Route(m *mux.Router)
}

type ProviderFactory func(ctx context.Context, backend string, key []byte) (Provider, error)

var providers = map[string]ProviderFactory{}

func Register(name string, factory ProviderFactory) {
	providers[name] = factory
}

func Init(ctx context.Context, r *mux.Router, backend string, key []byte) error {
	for k, v := range providers {
		p, err := v(ctx, backend, key)
		if err != nil {
			return err
		}
		p.Route(r.PathPrefix("/" + k).Subrouter())
	}
	return nil
}
