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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"

	cache2 "go.linka.cloud/artifact-registry/pkg/cache"
	"go.linka.cloud/artifact-registry/pkg/crypt/aes"
)

var cache = cache2.New()

type storage struct {
	host   string
	name   string
	rrepo  *remote.Repository
	ref    string
	repo   Repository
	key    string
	tmp    string
	aesKey []byte
}

func NewStorage(ctx context.Context, host, name string, repo Repository, aesKey []byte) (Storage, error) {
	name = host + "/" + strings.TrimSuffix(name, "/")
	ref := name + ":" + repo.Name()
	tmp, err := os.MkdirTemp(os.TempDir(), "lk-artifact-registry-"+repo.Name())
	if err != nil {
		return nil, err
	}
	r := &storage{
		host:   host,
		name:   name,
		repo:   repo,
		ref:    ref,
		tmp:    tmp,
		aesKey: aesKey,
	}
	defer func() {
		if err != nil {
			r.Close()
		}
	}()
	r.rrepo, err = r.newRepository(ctx, name)
	if err != nil {
		return nil, err
	}
	if err = r.fetchKey(ctx); err == nil {
		return r, nil
	}
	if !errors.Is(err, errdef.ErrNotFound) {
		return nil, err
	}
	if err = r.init(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *storage) Stat(ctx context.Context, file string) (ArtifactInfo, error) {
	desc, err := s.find(ctx, file)
	if err != nil {
		return nil, err
	}
	// TODO(adphi): retrieve version
	return &info{path: file, size: desc.Size, digest: desc.Digest, meta: desc.Data}, nil
}

func (s *storage) Open(ctx context.Context, file string) (io.ReadCloser, error) {
	desc, err := s.find(ctx, file)
	if err != nil {
		return nil, err
	}
	rd, err := s.rrepo.Blobs().Fetch(ctx, desc)
	if err != nil {
		return nil, err
	}
	return rd, nil
}

func (s *storage) Write(ctx context.Context, pkg Artifact) error {
	store, err := file.New(s.tmp)
	if err != nil {
		return err
	}
	pkgb, err := json.Marshal(pkg)
	if err != nil {
		return err
	}
	cfg := ocispec.Descriptor{
		MediaType: s.MediaTypeArtifactConfig(),
		Digest:    digest.FromBytes(pkgb),
		Size:      int64(len(pkgb)),
	}
	if err := store.Push(ctx, cfg, bytes.NewReader(pkgb)); err != nil {
		return err
	}
	layer := ocispec.Descriptor{
		MediaType: s.MediaTypeArtifactLayer(),
		Digest:    pkg.Digest(),
		Size:      pkg.Size(),
		Annotations: map[string]string{
			ocispec.AnnotationTitle: pkg.Path(),
		},
		Data: pkgb,
	}
	if err := store.Push(ctx, layer, pkg); err != nil {
		if errors.Is(err, file.ErrDuplicateName) {
			return fmt.Errorf("%s: %w", pkg.Path(), os.ErrExist)
		}
		return err
	}
	opts := oras.PackManifestOptions{
		ConfigDescriptor: &cfg,
		Layers:           []ocispec.Descriptor{layer},
	}
	img, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1_RC4, s.ArtefactTypeRegistry(), opts)
	if err != nil {
		return err
	}
	repo := s.name + "/" + pkg.Name()
	ref := strings.NewReplacer("~", "-", "+", "-").Replace(repo + ":" + pkg.Version())
	if err := store.Tag(ctx, img, img.Digest.String()); err != nil {
		return err
	}
	rrepo, err := s.newRepository(ctx, repo)
	if err != nil {
		return err
	}
	img, err = oras.Copy(ctx, store, img.Digest.String(), rrepo, ref, copts(repo))
	if err != nil {
		return err
	}
	m, err := s.manifest(ctx)
	if err != nil {
		return err
	}
	var ls []ocispec.Descriptor
	for _, v := range m.Layers {
		if v.Annotations[ocispec.AnnotationTitle] == pkg.Path() {
			logrus.Infof("updating layer %s (%s-", pkg.Path(), v.Digest)
			continue
		}
		ls = append(ls, v)
	}
	m.Layers = ls
	return s.updateIndex(ctx, store, m, []Artifact{pkg}, []ocispec.Descriptor{layer})
}

func (s *storage) Delete(ctx context.Context, name string) error {
	desc, err := s.find(ctx, name)
	if err != nil {
		return err
	}
	pkg, err := s.repo.Codec().Decode(desc.Data)
	if err != nil {
		return err
	}
	repo := s.name + "/" + pkg.Name()
	ref := strings.NewReplacer("~", "-", "+", "-").Replace(repo + ":" + pkg.Version())
	rrepo, err := s.newRepository(ctx, repo)
	if err != nil {
		return err
	}
	del := true
	pdesc, err := rrepo.Resolve(ctx, ref)
	if err != nil {
		if !errors.Is(err, errdef.ErrNotFound) {
			return err
		}
		del = false
	}
	if del {
		if err := rrepo.Delete(ctx, pdesc); err != nil {
			return err
		}
	}
	m, err := s.manifest(ctx)
	if err != nil {
		return err
	}
	var ls []ocispec.Descriptor
	for _, v := range m.Layers {
		if v.Annotations[ocispec.AnnotationTitle] == pkg.Path() {
			logrus.Infof("updating layer %s (%s-", pkg.Path(), v.Digest)
			continue
		}
		ls = append(ls, v)
	}
	m.Layers = ls
	store, err := file.New(s.tmp)
	if err != nil {
		return err
	}
	return s.updateIndex(ctx, store, m, nil, nil)
}

func (s *storage) ServeFile(w http.ResponseWriter, r *http.Request, path string) error {
	ctx := r.Context()
	desc, err := s.find(ctx, path)
	name := strings.Replace(path, "/", "-", -1)
	rd, err := s.rrepo.Blobs().Fetch(ctx, desc)
	if err != nil {
		return err
	}
	defer rd.Close()
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype != "" {
		w.Header().Set("Content-Type", ctype)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", desc.Size))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%s`, name))
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	_, err = io.Copy(w, rd)
	return err
}

func (s *storage) Key() string {
	return s.key
}

func (s *storage) Close() error {
	return os.RemoveAll(s.tmp)
}

func (s *storage) newRepository(ctx context.Context, name string) (*remote.Repository, error) {
	rrepo, err := remote.NewRepository(name)
	if err != nil {
		return nil, err
	}
	rrepo.Client = client(ctx, s.host)
	rrepo.PlainHTTP = plainHTTP
	return rrepo, nil
}

func (s *storage) updateIndex(ctx context.Context, store *file.Store, m ocispec.Manifest, pkgs []Artifact, layers []ocispec.Descriptor) error {
	pvn, pbn := s.repo.KeyNames()
	for i := range m.Layers {
		v := m.Layers[i]
		if n := v.Annotations[ocispec.AnnotationTitle]; n == pvn || n == pbn {
			layers = append(layers, v)
			continue
		}
		if v.MediaType != s.MediaTypeArtifactLayer() {
			continue
		}
		p, err := s.repo.Codec().Decode(v.Data)
		if err != nil {
			return err
		}
		pkgs = append(pkgs, p)
		layers = append(layers, v)
	}
	files, err := s.repo.Index(ctx, s.key, pkgs...)
	if err != nil {
		return err
	}
	opts := oras.PackManifestOptions{
		Layers: layers,
	}
	for _, v := range files {
		l := ocispec.Descriptor{
			MediaType: s.MediaTypeRegistryLayerMetadata(v.Name()),
			Digest:    v.Digest(),
			Size:      v.Size(),
			Annotations: map[string]string{
				ocispec.AnnotationTitle: v.Name(),
			},
		}
		if err := store.Push(ctx, l, v); err != nil {
			return err
		}
		opts.Layers = append(opts.Layers, l)
	}
	img, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1_RC4, s.ArtefactTypeRegistry(), opts)
	if err != nil {
		return err
	}
	if err := store.Tag(ctx, img, img.Digest.String()); err != nil {
		return err
	}
	// TODO(adphi): update only manifest
	img, err = oras.Copy(ctx, store, img.Digest.String(), s.rrepo, s.ref, copts(s.ref))
	if err != nil {
		return err
	}
	logrus.Infof("uploaded %s", s.ref)
	return nil
}

func (s *storage) init(ctx context.Context) error {
	store, err := file.New(s.tmp)
	if err != nil {
		return err
	}
	priv, pub, err := s.repo.GenerateKeypair()
	if err != nil {
		return err
	}
	s.key = priv
	enc, err := aes.Encrypt(s.aesKey, priv)
	if err != nil {
		return err
	}
	var opts oras.PackManifestOptions
	pvn, pbn := s.repo.KeyNames()
	for _, v := range []Artifact{NewFile(pvn, enc), NewFile(pbn, []byte(pub))} {
		l := ocispec.Descriptor{
			MediaType: s.MediaTypeRegistryLayerMetadata(v.Name()),
			Digest:    v.Digest(),
			Size:      v.Size(),
			Annotations: map[string]string{
				ocispec.AnnotationTitle: v.Path(),
			},
		}
		if err := store.Push(ctx, l, v); err != nil {
			return err
		}
		opts.Layers = append(opts.Layers, l)
	}
	img, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1_RC4, s.ArtefactTypeRegistry(), opts)
	if err != nil {
		return err
	}
	if err := store.Tag(ctx, img, img.Digest.String()); err != nil {
		return err
	}
	img, err = oras.Copy(ctx, store, img.Digest.String(), s.rrepo, s.ref, copts(s.ref))
	if err != nil {
		return err
	}
	logrus.Infof("storage inititalized %s", s.ref)
	return nil
}

func (s *storage) manifest(ctx context.Context) (m ocispec.Manifest, err error) {
	desc, err := s.rrepo.Resolve(ctx, s.repo.Name())
	if err != nil {
		return m, err
	}
	if v, ok := cache.Get(desc.Digest.String()); ok {
		// reset ttl
		cache.Set(desc.Digest.String(), v, cache2.WithTTL(5*time.Minute))
		return v.(ocispec.Manifest), nil
	}
	b, err := s.rrepo.Manifests().Fetch(ctx, desc)
	if err != nil {
		return m, err
	}
	defer b.Close()

	if err := json.NewDecoder(b).Decode(&m); err != nil {
		return m, err
	}
	cache.Set(desc.Digest.String(), m, cache2.WithTTL(5*time.Minute))
	return m, nil
}

func (s *storage) fetchKey(ctx context.Context) error {
	n, _ := s.repo.KeyNames()
	desc, err := s.find(ctx, n)
	if err != nil {
		return err
	}
	if v, ok := cache.Get(desc.Digest.String()); ok {
		s.key = v.(string)
		return nil
	}
	rd, err := s.rrepo.Blobs().Fetch(ctx, desc)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(rd)
	if err != nil {
		return err
	}
	priv, err := aes.Decrypt(s.aesKey, b)
	if err != nil {
		return err
	}
	s.key = string(priv)
	cache.Set(desc.Digest.String(), s.key)
	return nil
}

func (s *storage) find(ctx context.Context, file string) (ocispec.Descriptor, error) {
	m, err := s.manifest(ctx)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	for _, v := range m.Layers {
		if v.Annotations[ocispec.AnnotationTitle] == file {
			return v, nil
		}
	}
	return ocispec.Descriptor{}, fmt.Errorf("%s: %w", file, os.ErrNotExist)
}

func (s *storage) ArtefactTypeRegistry() string {
	return "application/vnd.lk.registry+" + s.repo.Name()
}
func (s *storage) MediaTypeArtifactConfig() string {
	return "application/vnd.lk.registry.config.v1." + s.repo.Name() + "+" + s.repo.Codec().Name()
}
func (s *storage) MediaTypeRegistryLayerMetadata(name string) string {
	if ext := filepath.Ext(name); ext != "" {
		return "application/vnd.lk.registry.metadata.layer.v1." + s.repo.Name() + "+" + ext
	}
	return "application/vnd.lk.registry.metadata.layer.v1." + s.repo.Name()
}
func (s *storage) MediaTypeArtifactLayer() string {
	return "application/vnd.lk.registry.layer.v1." + s.repo.Name()
}
