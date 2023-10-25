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
	"io"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

type Repository = registry.Repository

type RepositoryProxy interface {
	registry.ReferenceFetcher
	content.Fetcher
}

func NewRepository(ctx context.Context, reference string, opts ...Option) (Repository, error) {
	r, err := remote.NewRepository(reference)
	if err != nil {
		return nil, err
	}
	o := makeOptions(r.Reference.Host(), opts...)
	o.apply(ctx, r)
	var proxy registry.Repository
	if o.proxy.host != "" {
		ref, err := registry.ParseReference(reference)
		if err != nil {
			return nil, err
		}
		ref.Registry = o.proxy.host
		p, err := remote.NewRepository(ref.String())
		if err != nil {
			return nil, err
		}
		o.proxy.apply(ctx, p)
		proxy = p
	}
	return &repo{Repository: r, p: proxy}, nil
}

type repo struct {
	Repository
	p Repository
}

func (r *repo) Fetch(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error) {
	return r.maybeProxy().Fetch(ctx, target)
}

func (r *repo) FetchReference(ctx context.Context, reference string) (ocispec.Descriptor, io.ReadCloser, error) {
	return r.maybeProxy().FetchReference(ctx, reference)
}

func (r *repo) Blobs() BlobStore {
	b := blobs{BlobStore: r.Repository.Blobs()}
	if r.p != nil {
		b.p = r.p.Blobs()
	}
	return &b
}

func (r *repo) Manifests() ManifestStore {
	m := &manifests{ManifestStore: r.Repository.Manifests()}
	if r.p != nil {
		m.p = r.p.Manifests()
	}
	return m
}

func (r *repo) maybeProxy() RepositoryProxy {
	if r.p != nil {
		return r.p
	}
	return r.Repository
}
