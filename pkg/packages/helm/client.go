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

package helm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	hclient "go.linka.cloud/artifact-registry/pkg/http/client"
	"go.linka.cloud/artifact-registry/pkg/packages"
)

var _ packages.Client = (*client)(nil)

func NewClient(registry, repository string, opts ...hclient.Option) (packages.Client, error) {
	if registry == "" {
		return nil, fmt.Errorf("registry is required")
	}
	var base string
	if strings.HasPrefix(registry, Name+".") {
		base = fmt.Sprintf("%s/%s", registry, repository)
	} else {
		base = fmt.Sprintf("%s/%s/%s", registry, Name, repository)
	}
	return &client{
		c:          hclient.New(opts...),
		repository: repository,
		base:       strings.TrimSuffix(base, "/"),
	}, nil
}

type client struct {
	c          hclient.Client
	repository string
	base       string
}

func (c *client) Key(ctx context.Context) (string, error) {
	res, err := c.c.Get(ctx, c.path(RepositoryPublicKey))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *client) SetupScript(ctx context.Context) (string, error) {
	res, err := c.c.Get(ctx, c.path("setup"))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *client) Push(ctx context.Context, r io.Reader) error {
	_, err := c.c.Put(ctx, c.path("push"), r)
	return err
}

func (c *client) Pull(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	res, err := c.c.Get(ctx, c.path(path))
	if err != nil {
		return nil, 0, err
	}
	return res.Body, res.ContentLength, nil
}

func (c *client) Delete(ctx context.Context, path string) error {
	_, err := c.c.Delete(ctx, c.path(path))
	return err
}

func (c *client) path(parts ...string) string {
	return fmt.Sprintf("%s/%s", c.base, strings.Join(parts, "/"))
}
