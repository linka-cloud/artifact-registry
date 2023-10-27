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
	"fmt"
	"strings"

	hclient "go.linka.cloud/artifact-registry/pkg/http/client"
	"go.linka.cloud/artifact-registry/pkg/packages"
	"go.linka.cloud/artifact-registry/pkg/packages/apk"
	"go.linka.cloud/artifact-registry/pkg/packages/deb"
	"go.linka.cloud/artifact-registry/pkg/packages/helm"
	"go.linka.cloud/artifact-registry/pkg/packages/rpm"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

type Client interface {
	Login(ctx context.Context) error
	Repositories(ctx context.Context) ([]*Repository, error)
	Packages(ctx context.Context, typ string) ([]storage.Artifact, error)
}

func NewClient(registry string, repo string, opts ...hclient.Option) (Client, error) {
	if registry == "" {
		return nil, fmt.Errorf("registry is required")
	}
	var typ string
	for _, v := range packages.Names() {
		if strings.HasPrefix(registry, v+".") {
			typ = v
			break
		}
	}
	return &client{
		c:        hclient.New(opts...),
		registry: registry,
		repo:     repo,
		typ:      typ,
	}, nil
}

type client struct {
	c        hclient.Client
	registry string
	repo     string
	typ      string
}

func (c *client) Login(ctx context.Context) error {
	_, err := c.c.Get(ctx, ie(c.repo != "", c.url("_auth", c.repo, "login"), c.url("_auth", "login")))
	return err
}

func (c *client) Repositories(ctx context.Context) ([]*Repository, error) {
	res, err := c.c.Get(ctx, ie(c.repo != "", c.url("_repositories", c.repo), c.url("_repositories")))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var repos []*Repository
	return repos, json.NewDecoder(res.Body).Decode(&repos)
}

func (c *client) Packages(ctx context.Context, typ string) ([]storage.Artifact, error) {
	if typ == "" && c.typ == "" {
		return nil, fmt.Errorf("package type is required when not included in subdomain")
	}
	if typ == c.typ {
		typ = ""
	}
	b := ie(typ != "", "_packages/"+typ, "_packages")
	res, err := c.c.Get(ctx, ie(c.repo != "", c.url(b, c.repo), c.url(b)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	switch ie(typ != "", typ, c.typ) {
	case apk.Name:
		var p []*apk.Package
		err := json.NewDecoder(res.Body).Decode(&p)
		return storage.AsArtifact(p), err
	case deb.Name:
		var p []*deb.Package
		err := json.NewDecoder(res.Body).Decode(&p)
		return storage.AsArtifact(p), err
	case rpm.Name:
		var p []*rpm.Package
		err := json.NewDecoder(res.Body).Decode(&p)
		return storage.AsArtifact(p), err
	case helm.Name:
		var p []*helm.Package
		err := json.NewDecoder(res.Body).Decode(&p)
		return storage.AsArtifact(p), err
	default:
		return nil, fmt.Errorf("unexpected package type %q", typ)
	}
}

func (c *client) url(parts ...string) string {
	return fmt.Sprintf("%s/%s", c.registry, strings.Join(parts, "/"))
}

func ie[T any](c bool, a T, b T) T {
	if c {
		return a
	}
	return b
}
