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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hclient "go.linka.cloud/artifact-registry/pkg/http/client"
	"go.linka.cloud/artifact-registry/pkg/packages"
)

func TestSetupLocal(t *testing.T) {
	// Create a mock client that returns a fake key
	inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("FAKEKEY")),
		}, nil
	})))
	t.Run("no repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("example.com", "", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/apt/sources.list.d/example-com.list")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://example.com/deb stable main")

		key, err := afero.ReadFile(fs, "/etc/apt/trusted.gpg.d/example-com.asc")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("subdomain without repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("deb.example.com", "", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/apt/sources.list.d/example-com.list")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://deb.example.com stable main")

		key, err := afero.ReadFile(fs, "/etc/apt/trusted.gpg.d/example-com.asc")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/apt/sources.list.d/my-repo.list")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://example.com/deb/my-repo stable main")

		key, err := afero.ReadFile(fs, "/etc/apt/trusted.gpg.d/my-repo.asc")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("subdomain with repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("deb.example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/apt/sources.list.d/my-repo.list")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://deb.example.com/my-repo stable main")

		key, err := afero.ReadFile(fs, "/etc/apt/trusted.gpg.d/my-repo.asc")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("with credentials", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = hclient.New(hclient.WithBasicAuth("user", "pass"), hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			u, p, ok := req.BasicAuth()
			require.True(t, ok)
			assert.Equal(t, "user", u)
			assert.Equal(t, "pass", p)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("FAKEKEY")),
			}, nil
		})))

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/apt/sources.list.d/my-repo.list")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://example.com/deb/my-repo stable main")
		contents, err = afero.ReadFile(fs, "/etc/apt/auth.conf.d/my-repo.conf")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "machine https://example.com/deb/my-repo/stable/main login user password pass")

		key, err := afero.ReadFile(fs, "/etc/apt/trusted.gpg.d/my-repo.asc")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("already configured", func(t *testing.T) {
		// Mock filesystem with existing config
		fs = afero.NewMemMapFs()
		afero.WriteFile(fs, "/etc/apt/sources.list.d/my-repo.list", []byte("deb https://example.com/deb stable main"), 0644)

		// Create a client
		c, err := NewClient("example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		assert.Equal(t, packages.ErrAlreadyConfigured, err)
	})

	t.Run("force update", func(t *testing.T) {
		// Mock filesystem with existing config
		fs = afero.NewMemMapFs()
		afero.WriteFile(fs, "/etc/apt/sources.list.d/my-repo.list", []byte("deb https://example.com/deb stable main"), 0644)

		// Create a client
		c, err := NewClient("example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal with force=true
		err = c.SetupLocal(context.Background(), true)

		// Verify no error
		require.NoError(t, err)
	})

	t.Run("get key error", func(t *testing.T) {
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Status:     "Internal Server Error",
			}, nil
		})))

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to download repository key")
	})
}
