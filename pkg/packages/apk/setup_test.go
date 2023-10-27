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

package apk

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
	// Create a mock HTTP client that returns a fake key
	inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				"Content-Disposition": {"attachment; filename=key.pub"},
			},
			Body: io.NopCloser(strings.NewReader("FAKEKEY")),
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
		contents, err := afero.ReadFile(fs, "/etc/apk/repositories")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://example.com/apk/stable/main")

		key, err := afero.ReadFile(fs, "/etc/apk/keys/key.pub")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("subdomain without repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("apk.example.com", "", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/apk/repositories")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://apk.example.com/stable/main")

		key, err := afero.ReadFile(fs, "/etc/apk/keys/key.pub")
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
		contents, err := afero.ReadFile(fs, "/etc/apk/repositories")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://example.com/apk/my-repo/stable/main")

		key, err := afero.ReadFile(fs, "/etc/apk/keys/key.pub")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("subdomain with repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("apk.example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/apk/repositories")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://apk.example.com/my-repo/stable/main")

		key, err := afero.ReadFile(fs, "/etc/apk/keys/key.pub")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("with credentials", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a mock HTTP client that returns a fake key
		inner := hclient.New(hclient.WithBasicAuth("user", "pass"), hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Disposition": {"attachment; filename=key.pub"},
				},
				Body: io.NopCloser(strings.NewReader("FAKEKEY")),
			}, nil
		})))

		// Create a client
		c, err := NewClient("example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/apk/repositories")
		require.NoError(t, err)
		assert.Contains(t, string(contents), "https://user:pass@example.com/apk/my-repo/stable/main")

		key, err := afero.ReadFile(fs, "/etc/apk/keys/key.pub")
		require.NoError(t, err)
		assert.Equal(t, "FAKEKEY", string(key))
	})

	t.Run("already configured", func(t *testing.T) {
		// Mock filesystem with existing config
		fs = afero.NewMemMapFs()
		afero.WriteFile(fs, "/etc/apk/repositories", []byte("https://example.com/apk/stable/main"), 0644)

		// Create client
		c, err := NewClient("example.com", "", "stable", "main")
		require.NoError(t, err)

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		assert.Equal(t, packages.ErrAlreadyConfigured, err)
	})

	t.Run("force update", func(t *testing.T) {
		// Mock filesystem with existing config
		fs = afero.NewMemMapFs()
		afero.WriteFile(fs, "/etc/apk/repositories", []byte("https://example.com/apk/stable/main"), 0644)

		// Create a mock HTTP client that returns a fake key
		inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Content-Disposition": {"attachment; filename=key.pub"},
				},
				Body: io.NopCloser(strings.NewReader("FAKEKEY")),
			}, nil
		})))

		// Create a client
		c, err := NewClient("apk.example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal with force=true
		err = c.SetupLocal(context.Background(), true)

		// Verify no error
		require.NoError(t, err)
	})

	t.Run("get key error", func(t *testing.T) {
		fs = afero.NewMemMapFs()
		// Mock client that returns error
		inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Status:     "Internal Server Error",
			}, nil
		})))

		// Create a client
		c, err := NewClient("apk.example.com", "my-repo", "stable", "main")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get repository key")
	})
}
