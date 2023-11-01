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

package deb

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hclient "go.linka.cloud/artifact-registry/pkg/http/client"
)

var ErrSkip = errors.New("skip")

func TestClientURL(t *testing.T) {
	type test struct {
		name       string
		registry   string
		repository string
		branch     string
		repo       string
		fn         func(ctx context.Context, c *client) error
		url        string
		wantErr    bool
	}
	tests := []test{
		{
			name:       "invalid registry",
			repository: "my-repo",
			branch:     "stable",
			repo:       "main",
			wantErr:    true,
		},
		{
			name:     "without repo",
			registry: "example.org",
			branch:   "stable",
			repo:     "main",
		},
		{
			name:       "with repo",
			registry:   "example.org",
			repository: "my-repo",
			branch:     "stable",
			repo:       "main",
		},
		{
			name:     "without repo (subpath)",
			registry: "example.org",
			branch:   "stable",
			repo:     "main",
			url:      "https://example.org/deb/" + RepositoryPublicKey,
			fn: func(ctx context.Context, c *client) error {
				_, err := c.Key(ctx)
				return err
			},
		},
		{
			name:     "without repo (subdomain)",
			registry: "deb.example.org",
			branch:   "stable",
			repo:     "main",
			url:      "https://deb.example.org/" + RepositoryPublicKey,
			fn: func(ctx context.Context, c *client) error {
				_, err := c.Key(ctx)
				return err
			},
		},
		{
			name:     "without repo (subdomain other type)",
			registry: "apk.example.org",
			branch:   "stable",
			repo:     "main",
			url:      "https://apk.example.org/deb/" + RepositoryPublicKey,
			fn: func(ctx context.Context, c *client) error {
				_, err := c.Key(ctx)
				return err
			},
		},
		{
			name:       "with repo (subpath)",
			registry:   "example.org",
			repository: "my-repo",
			branch:     "stable",
			repo:       "main",
			url:        "https://example.org/deb/my-repo/" + RepositoryPublicKey,
			fn: func(ctx context.Context, c *client) error {
				_, err := c.Key(ctx)
				return err
			},
		},
		{
			name:       "with repo (subdomain)",
			registry:   "deb.example.org",
			repository: "my-repo",
			branch:     "stable",
			repo:       "main",
			url:        "https://deb.example.org/my-repo/" + RepositoryPublicKey,
			fn: func(ctx context.Context, c *client) error {
				_, err := c.Key(ctx)
				return err
			},
		},
		{
			name:       "with repo (subdomain other type)",
			registry:   "apk.example.org",
			repository: "my-repo",
			branch:     "stable",
			repo:       "main",
			url:        "https://apk.example.org/deb/my-repo/" + RepositoryPublicKey,
			fn: func(ctx context.Context, c *client) error {
				_, err := c.Key(ctx)
				return err
			},
		},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			c, err := NewClient(v.registry, v.repository, v.branch, v.repo)
			if v.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if v.fn == nil {
				return
			}
			c.(*client).c = hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, v.url, r.URL.String())
				return nil, ErrSkip
			})))
			err = v.fn(ctx, c.(*client))
			assert.ErrorIs(t, err, ErrSkip)
		})
	}
}
