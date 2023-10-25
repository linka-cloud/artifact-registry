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
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/auth/htpasswd"
	_ "github.com/distribution/distribution/v3/registry/auth/token"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.linka.cloud/grpc-toolkit/logger"

	"go.linka.cloud/artifact-registry/pkg/auth"
	"go.linka.cloud/artifact-registry/pkg/codec"
	"go.linka.cloud/artifact-registry/pkg/crypt/aes"
	registry2 "go.linka.cloud/artifact-registry/pkg/registry"
	"go.linka.cloud/artifact-registry/pkg/slices"
)

const (
	addr = "localhost:5555"
	repo = "test"
)

func newRegistry(t *testing.T, ctx context.Context, auth map[string]configuration.Parameters) *registry.Registry {
	t.Helper()
	config := &configuration.Configuration{}
	config.Log.AccessLog.Disabled = true
	config.Log.Level = configuration.Loglevel(logger.FatalLevel.String())
	config.HTTP.Addr = addr
	config.Auth = auth
	config.Storage = map[string]configuration.Parameters{"inmemory": map[string]interface{}{}}
	reg, err := registry.NewRegistry(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

var _ Artifact = (*mockArtifact)(nil)

type mockArtifact struct {
	S string `json:"s"`
	r io.Reader
}

func newMockArtifact(s string) *mockArtifact {
	return &mockArtifact{S: s}
}

func (m *mockArtifact) Read(p []byte) (n int, err error) {
	if m.r == nil {
		m.r = strings.NewReader(m.S)
	}
	return m.r.Read(p)
}

func (m *mockArtifact) Close() error {
	return nil
}

func (m *mockArtifact) Name() string {
	return m.S
}

func (m *mockArtifact) Path() string {
	return m.S
}

func (m *mockArtifact) Arch() string {
	return "noarch"
}

func (m *mockArtifact) Version() string {
	return "none"
}

func (m *mockArtifact) Size() int64 {
	return int64(len(m.S))
}

func (m *mockArtifact) Digest() digest.Digest {
	return digest.FromString(m.S)
}

var _ Repository = (*mockRepository)(nil)

type mockRepository struct{}

func (m *mockRepository) Index(_ context.Context, _ string, artifacts ...Artifact) ([]Artifact, error) {
	return []Artifact{NewFile("index.txt", []byte(strings.Join(slices.Map(artifacts, func(a Artifact) string {
		return a.Name()
	}), "\n")))}, nil
}

func (m *mockRepository) GenerateKeypair() (string, string, error) {
	return "private", "public", nil
}

func (m *mockRepository) KeyNames() (string, string) {
	return "repository.key", "repository.pub"
}

func (m *mockRepository) Codec() Codec {
	return codec.Funcs[Artifact]{
		Format: "json",
		EncodeFunc: func(v Artifact) ([]byte, error) {
			return json.Marshal(v)
		},
		DecodeFunc: func(b []byte) (Artifact, error) {
			var m mockArtifact
			if err := json.Unmarshal(b, &m); err != nil {
				return nil, err
			}
			return &m, nil
		},
	}
}

func (m *mockRepository) Name() string {
	return "mock"
}

type mockAuth string

func (m mockAuth) BasicAuth() (string, string, bool) {
	return string(m), string(m), true
}

type test struct {
	name string
	fn   func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository)
}

func TestStorage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmp, err := os.MkdirTemp("", "lkar-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmp)
	htpasswd := filepath.Join(tmp, "htpasswd")
	require.NoError(t, os.WriteFile(htpasswd, []byte("test:$2y$05$t/k5UmvFFg0I19w7jYUpBe1z4rvw1VTsYHv9d8I9jeXssE1UA9m7u"), os.ModePerm))

	reg := newRegistry(t, ctx, map[string]configuration.Parameters{"htpasswd": map[string]interface{}{"realm": "test", "path": htpasswd}})
	go reg.ListenAndServe()

	k := sha256.Sum256([]byte("test"))
	ctx = WithOptions(ctx, WithHost(addr), WithKey(k[:]), WithRegistryOptions(registry2.WithPlainHTTP()))

	var s *storage
	defer func() {
		if s != nil {
			s.Close()
		}
	}()

	time.Sleep(time.Second)

	t.Run("storage requires authentication", func(t *testing.T) {
		_, err := NewStorage(ctx, repo, &mockRepository{})
		// require.ErrorContains(t, err, "Unauthorized")
		require.Error(t, err)
	})

	t.Run("storage forwards authentication", func(t *testing.T) {
		ctx = auth.Context(ctx, mockAuth("test"))
		v, err := NewStorage(ctx, repo, &mockRepository{})
		require.NoError(t, err)
		s = v.(*storage)
	})

	r, err := Options(ctx).NewRepository(ctx, addr+"/"+repo)
	require.NoError(t, err)

	tests := []test{
		{
			name: "new registry is not initialized",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				assert.Empty(t, s.Key())
				_, err := r.Resolve(ctx, "mock")
				assert.ErrorContains(t, err, "not found")
			},
		},
		{
			name: "write initializes registry with keys",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				require.NoError(t, s.Write(ctx, newMockArtifact("test.txt")))
				assert.Equal(t, "private", s.Key())
				desc, err := s.find(ctx, "repository.key")
				require.NoError(t, err)
				assert.Equal(t, s.MediaTypeRegistryLayerMetadata("repository.key"), desc.MediaType)
			},
		},
		{
			name: "private key is encrypted",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				desc, err := s.find(ctx, "repository.key")
				require.NoError(t, err)
				rd, err := s.rrepo.Blobs().Fetch(ctx, desc)
				require.NoError(t, err)
				defer rd.Close()
				b, err := io.ReadAll(rd)
				assert.NotEqual(t, "private", string(b))
				kb, err := aes.Decrypt(k[:], b)
				require.NoError(t, err)
				assert.Equal(t, "private", string(kb))
			},
		},
		{
			name: "cannot read private key",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				_, err := s.Open(ctx, "repository.key")
				assert.ErrorIs(t, err, os.ErrNotExist)
			},
		},
		{
			name: "cannot delete public and private key",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				assert.ErrorIs(t, s.Delete(ctx, "repository.key"), os.ErrNotExist)
				assert.ErrorIs(t, s.Delete(ctx, "repository.pub"), os.ErrNotExist)
			},
		},
		{
			name: "write store artifacts, index and keys in the manifest",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				m, err := s.manifest(ctx)
				require.NoError(t, err)
				assert.Equal(t, s.ArtefactTypeRegistry(), m.ArtifactType)
				// private & public key & artifact & index
				require.Len(t, m.Layers, 4)
				for i, v := range []string{
					"test.txt",
					"repository.key",
					"repository.pub",
					"index.txt",
				} {
					assert.Equal(t, v, m.Layers[i].Annotations[ocispec.AnnotationTitle])
				}
				rc, err := s.Open(ctx, "test.txt")
				require.NoError(t, err)
				defer rc.Close()
				b, err := io.ReadAll(rc)
				require.NoError(t, err)
				assert.Equal(t, "test.txt", string(b))

				rc, err = s.Open(ctx, "index.txt")
				require.NoError(t, err)
				defer rc.Close()

				b, err = io.ReadAll(rc)
				require.NoError(t, err)
				assert.Equal(t, "test.txt", string(b))
			},
		},
		{
			name: "artifact metadata is stored in the manifest",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				desc, err := s.find(ctx, "test.txt")
				require.NoError(t, err)
				assert.Equal(t, s.MediaTypeArtifactLayer(), desc.MediaType)
				assert.Equal(t, `{"s":"test.txt"}`, string(desc.Data))
			},
		},
		{
			name: "write updates the index",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				require.NoError(t, s.Write(ctx, newMockArtifact("test2.txt")))
				rc, err := s.Open(ctx, "index.txt")
				require.NoError(t, err)
				defer rc.Close()
				b, err := io.ReadAll(rc)
				require.NoError(t, err)
				assert.Equal(t, "test2.txt\ntest.txt", string(b))
			},
		},
		{
			name: "written artifacts are readable",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				rc, err := s.Open(ctx, "test2.txt")
				require.NoError(t, err)
				defer rc.Close()
				b, err := io.ReadAll(rc)
				require.NoError(t, err)
				assert.Equal(t, "test2.txt", string(b))
			},
		},
		{
			name: "delete removes the artifact from the index",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				require.NoError(t, s.Delete(ctx, "test.txt"))
				rc, err := s.Open(ctx, "index.txt")
				require.NoError(t, err)
				defer rc.Close()
				b, err := io.ReadAll(rc)
				require.NoError(t, err)
				assert.Equal(t, "test2.txt", string(b))
			},
		},
		{
			name: "delete removes the artifact from the manifest",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				_, err := s.find(ctx, "test.txt")
				assert.ErrorIs(t, err, os.ErrNotExist)
			},
		},
		{
			name: "deleted file is not readable",
			fn: func(t *testing.T, ctx context.Context, s *storage, reg registry2.Repository) {
				_, err := s.Open(ctx, "test.txt")
				assert.ErrorIs(t, err, os.ErrNotExist)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, ctx, s, r)
		})
	}
}
