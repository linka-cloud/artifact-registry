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

package rpm

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hclient "go.linka.cloud/artifact-registry/pkg/http/client"
	"go.linka.cloud/artifact-registry/pkg/packages"
)

func TestSetupLocal(t *testing.T) {
	// Create a mock HTTP client that returns a fake key
	repo := dedent.Dedent(`[example-com]
		name=example-com
		baseurl=https://example.com/rpm
		enabled=1
		gpgcheck=1
		gpgkey=https://example.com/rpm/key.pub
	`)
	inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(repo)),
		}, nil
	})))

	t.Run("no repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		// Create a client
		c, err := NewClient("example.com", "")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/yum.repos.d/example-com.repo")
		require.NoError(t, err)
		assert.Contains(t, string(contents), repo)
	})

	t.Run("subdomain without repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		repo := dedent.Dedent(`[example-com]
			name=example-com
			baseurl=https://rpm.example.com
			enabled=1
			gpgcheck=1
			gpgkey=https://repo.example.com/key.pub
		`)

		// Create a mock HTTP client that returns a fake key
		inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(repo)),
			}, nil
		})))

		// Create a client
		c, err := NewClient("rpm.example.com", "")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/yum.repos.d/example-com.repo")
		require.NoError(t, err)
		assert.Contains(t, string(contents), repo)
	})

	t.Run("repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		repo := dedent.Dedent(`[my-repo]
			name=my-repo
			baseurl=https://example.com/rpm/my-repo
			enabled=1
			gpgcheck=1
			gpgkey=https://example.com/rpm/my-repo/key.pub
		`)
		// Create a mock HTTP client that returns a fake key
		inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(repo)),
			}, nil
		})))

		// Create a client
		c, err := NewClient("example.com", "my-repo")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/yum.repos.d/my-repo.repo")
		require.NoError(t, err)
		assert.Contains(t, string(contents), repo)
	})

	t.Run("subdomain with repo name", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		repo := dedent.Dedent(`[my-repo]
			name=my-repo
			baseurl=https://rpm.example.com
			enabled=1
			gpgcheck=1
			gpgkey=https://rpm.example.com/key.pub
		`)

		// Create a mock HTTP client that returns a fake key
		inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(repo)),
			}, nil
		})))

		// Create a client
		c, err := NewClient("rpm.example.com", "my-repo")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/yum.repos.d/my-repo.repo")
		require.NoError(t, err)
		assert.Contains(t, string(contents), repo)
	})

	t.Run("with credentials", func(t *testing.T) {
		// Create a mock filesystem
		fs = afero.NewMemMapFs()

		repo := dedent.Dedent(`[example-com]
			name=example-com
			baseurl=https://example.com/rpm
			enabled=1
			gpgcheck=1
			gpgkey=https://example.com/rpm/key.pub
			username=user
			password=pass
		`)

		// Create a mock HTTP client that returns a fake key
		inner := hclient.New(hclient.WithBasicAuth("user", "pass"), hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(repo)),
			}, nil
		})))

		// Create a client
		c, err := NewClient("example.com", "my-repo")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify
		require.NoError(t, err)
		contents, err := afero.ReadFile(fs, "/etc/yum.repos.d/my-repo.repo")
		require.NoError(t, err)
		assert.Contains(t, string(contents), repo)
	})

	t.Run("already configured", func(t *testing.T) {
		// Mock filesystem with existing config
		fs = afero.NewMemMapFs()
		afero.WriteFile(fs, "/etc/yum.repos.d/my-repo.repo", []byte(""), 0644)

		// Create a mock HTTP client that returns a fake key
		inner := hclient.New(hclient.WithBasicAuth("user", "pass"), hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("FAKEKEY")),
			}, nil
		})))

		// Create a client
		c, err := NewClient("example.com", "my-repo")
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
		repo := dedent.Dedent(`[my-repo]
			name=my-repo
			baseurl=https://example.com/rpm
			enabled=1
			gpgcheck=1
			gpgkey=https://example.com/rpm/key.pub
		`)
		afero.WriteFile(fs, "/etc/apt/sources.list.d/my-repo.list", []byte("rpm https://example.com/rpm stable main"), 0644)

		// Create a mock HTTP client that returns a fake key
		inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(repo)),
			}, nil
		})))

		// Create a client
		c, err := NewClient("example.com", "my-repo")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal with force=true
		err = c.SetupLocal(context.Background(), true)

		// Verify no error
		require.NoError(t, err)
	})

	t.Run("get repo error", func(t *testing.T) {
		fs = afero.NewMemMapFs()
		// Mock client that returns error
		inner := hclient.New(hclient.WithTransport(hclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Status:     "Internal Server Error",
			}, nil
		})))

		// Create a client
		c, err := NewClient("example.com", "my-repo")
		require.NoError(t, err)
		c.(*client).c = inner

		// Call SetupLocal
		err = c.SetupLocal(context.Background(), false)

		// Verify error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get repository definition")
	})
}
