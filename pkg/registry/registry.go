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

package registry

import (
	"context"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

type Registry = registry.Registry

func NewRegistry(ctx context.Context, name string, opts ...Option) (Registry, error) {
	r, err := remote.NewRegistry(name)
	if err != nil {
		return nil, err
	}
	o := makeOptions(r.RepositoryOptions.Reference.Host(), opts...)
	o.apply(ctx, (*remote.Repository)(&r.RepositoryOptions))
	var proxy Registry
	if o.proxy.host != "" {
		p, err := remote.NewRegistry(o.proxy.host)
		if err != nil {
			return nil, err
		}
		o.proxy.apply(ctx, (*remote.Repository)(&p.RepositoryOptions))
		proxy = p
	}
	return &reg{r: r, p: proxy}, nil
}

type reg struct {
	r Registry
	p Registry
}

func (r *reg) Repositories(ctx context.Context, last string, fn func(repos []string) error) error {
	return r.r.Repositories(ctx, last, fn)
}

func (r *reg) Repository(ctx context.Context, name string) (Repository, error) {
	rep, err := r.r.Repository(ctx, name)
	if err != nil {
		return nil, err
	}
	var proxy Repository
	if r.p != nil {
		proxy, err = r.p.Repository(ctx, name)
		if err != nil {
			return nil, err
		}
	}
	return &repo{Repository: rep, p: proxy}, nil
}
