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
	"context"

	"go.linka.cloud/artifact-registry/pkg/registry"
)

type optionsKey struct{}

func WithOptions(ctx context.Context, opts ...Option) context.Context {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	return context.WithValue(ctx, optionsKey{}, o)
}

func Options(ctx context.Context) options {
	o, _ := ctx.Value(optionsKey{}).(options)
	return o
}

type options struct {
	host         string
	key          []byte
	repo         string
	artifactTags bool
	ropts        []registry.Option
}

func (o options) Host() string {
	return o.host
}

func (o options) Repo() string {
	return o.repo
}

func (o options) Key() []byte {
	return o.key
}

func (o options) NewRegistry(ctx context.Context) (registry.Registry, error) {
	return registry.NewRegistry(ctx, o.host, o.ropts...)
}

func (o options) NewRepository(ctx context.Context, name string) (registry.Repository, error) {
	return registry.NewRepository(ctx, name, o.ropts...)
}

type Option func(o *options)

func WithHost(host string) Option {
	return func(o *options) {
		o.host = host
	}
}

func WithKey(key []byte) Option {
	return func(o *options) {
		o.key = key
	}
}

func WithRepo(repo string) Option {
	return func(o *options) {
		o.repo = repo
	}
}

func WithArtifactTags() Option {
	return func(o *options) {
		o.artifactTags = true
	}
}

func WithRegistryOptions(opts ...registry.Option) Option {
	return func(o *options) {
		o.ropts = opts
	}
}
