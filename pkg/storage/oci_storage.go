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

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.linka.cloud/grpc-toolkit/logger"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/errdef"

	"go.linka.cloud/artifact-registry/pkg/cache"
	"go.linka.cloud/artifact-registry/pkg/crypt/aes"
	"go.linka.cloud/artifact-registry/pkg/mutex"
	"go.linka.cloud/artifact-registry/pkg/registry"
	"go.linka.cloud/artifact-registry/pkg/slices"
)

// global mutex to prevent concurrent access to the same storage
var lock = mutex.New()

type storage struct {
	opts  options
	name  string
	rrepo registry.Repository
	ref   string
	repo  Repository
	key   string
	tmp   string
}

func NewStorage(ctx context.Context, name string, repo Repository) (Storage, error) {
	opts := Options(ctx)
	if name == "" {
		if opts.repo == "" {
			return nil, errors.New("repository name is required")
		}
		name = opts.repo
	}
	rname := opts.host + "/" + strings.TrimSuffix(name, "/")
	ref := rname + ":" + repo.Name()
	tmp, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("lk-artifact-registry-%s-", repo.Name()))
	if err != nil {
		return nil, err
	}
	r := &storage{
		name: rname,
		repo: repo,
		ref:  ref,
		tmp:  tmp,
		opts: opts,
	}
	defer func() {
		if err != nil {
			r.Close()
		}
	}()
	r.rrepo, err = opts.NewRepository(ctx, rname)
	if err != nil {
		return nil, err
	}
	if err = r.fetchKey(ctx); err != nil {
		if !errors.Is(err, errdef.ErrNotFound) {
			return nil, err
		}
		err = nil
	}
	return r, nil
}

func (s *storage) Stat(ctx context.Context, file string) (ArtifactInfo, error) {
	logger.C(ctx).Infof("stat %s", file)
	desc, err := s.find(ctx, file)
	if err != nil {
		return nil, err
	}
	// TODO(adphi): retrieve version and maybe arch from manifest
	return &info{path: file, size: desc.Size, digest: desc.Digest, meta: desc.Data}, nil
}

func (s *storage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	logger.C(ctx).Infof("opening %s", path)
	if k, _ := s.repo.KeyNames(); path == k || path == "" {
		return nil, fmt.Errorf("%s: %w", path, os.ErrNotExist)
	}
	logger.C(ctx).Infof("downloading %s", path)
	desc, err := s.find(ctx, path)
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
	log := logger.C(ctx).WithField("artifact", pkg.Name())
	ctx = logger.Set(ctx, log)

	if err := s.Init(ctx); err != nil {
		return err
	}

	s.lock(ctx)
	defer s.unlock(ctx)

	log.Infof("uploading %s", pkg.Path())
	if prv, pb := s.repo.KeyNames(); pkg.Path() == prv || pkg.Path() == pb {
		return fmt.Errorf("%s: %w", pkg.Path(), os.ErrExist)
	}
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
	if s.opts.artifactTags {
		repo := s.artifactName(pkg)
		ref := strings.NewReplacer("~", "-", "+", "-").Replace(repo + ":" + defaults(pkg.Version(), "latest"))
		log.Infof("tagging artifact %s", ref)
		if err := store.Tag(ctx, img, img.Digest.String()); err != nil {
			return err
		}
		rrepo, err := s.opts.NewRepository(ctx, repo)
		if err != nil {
			return err
		}
		img, err = oras.Copy(ctx, store, img.Digest.String(), rrepo, ref, copts(repo))
		if err != nil {
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
			logger.C(ctx).Infof("updating layer %s (%s)", pkg.Path(), v.Digest)
			continue
		}
		ls = append(ls, v)
	}
	m.Layers = ls
	return s.updateIndex(ctx, store, m, []Artifact{pkg}, []ocispec.Descriptor{layer})
}

func (s *storage) Delete(ctx context.Context, name string) error {
	logger.C(ctx).Infof("deleting %s", name)
	s.lock(ctx)
	defer s.unlock(ctx)
	if prv, pb := s.repo.KeyNames(); name == prv || name == pb {
		return fmt.Errorf("%s: %w", name, os.ErrNotExist)
	}
	desc, err := s.find(ctx, name)
	if err != nil {
		return err
	}
	pkg, err := s.repo.Codec().Decode(desc.Data)
	if err != nil {
		return err
	}
	if s.opts.artifactTags {
		repo := s.artifactName(pkg)
		ref := strings.NewReplacer("~", "-", "+", "-").Replace(repo + ":" + defaults(pkg.Version(), "latest"))
		rrepo, err := s.opts.NewRepository(ctx, repo)
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
	}
	m, err := s.manifest(ctx)
	if err != nil {
		return err
	}
	var ls []ocispec.Descriptor
	for _, v := range m.Layers {
		if v.Annotations[ocispec.AnnotationTitle] == pkg.Path() {
			logger.C(ctx).Infof("updating layer %s (%s-", pkg.Path(), v.Digest)
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

func (s *storage) Artifacts(ctx context.Context) ([]Artifact, error) {
	logger.C(ctx).Infof("listing artifacts")
	m, err := s.manifest(ctx)
	if err != nil {
		return nil, err
	}
	var out []Artifact
	for _, v := range m.Layers {
		if v.MediaType != s.MediaTypeArtifactLayer() {
			continue
		}
		p, err := s.repo.Codec().Decode(v.Data)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *storage) ServeFile(w http.ResponseWriter, r *http.Request, path string) error {
	if k, _ := s.repo.KeyNames(); path == k || path == "" {
		return fmt.Errorf("%s: %w", path, os.ErrNotExist)
	}
	ctx := r.Context()

	logger.C(ctx).Infof("serving %s", path)
	desc, err := s.find(ctx, path)
	if err != nil {
		return err
	}
	name := filepath.Base(path)
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

func (s *storage) Size(ctx context.Context) (int64, error) {
	logger.C(ctx).Infof("computing storage size")
	m, err := s.manifest(ctx)
	if err != nil {
		return 0, err
	}
	var size int64
	l := make(map[string]struct{})
	for _, v := range m.Layers {
		if _, ok := l[v.Digest.String()]; ok {
			continue
		}
		size += v.Size
		l[v.Digest.String()] = struct{}{}
	}
	return size, nil
}

func (s *storage) Key() string {
	return s.key
}

func (s *storage) Close() error {
	return os.RemoveAll(s.tmp)
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
	logger.C(ctx).Infof("updating index")
	files, err := s.repo.Index(ctx, s.key, pkgs...)
	if err != nil {
		return err
	}
	i := make(map[string]string)
	for _, v := range append(pkgs, files...) {
		i[v.Path()] = v.Digest().String()
	}
	ib, err := json.Marshal(i)
	if err != nil {
		return fmt.Errorf("failed to marshal packages: %w", err)
	}
	cfg := ocispec.Descriptor{
		MediaType: s.MediaTypeIndexConfig(),
		Digest:    digest.FromBytes(ib),
		Size:      int64(len(ib)),
	}
	if err := store.Push(ctx, cfg, bytes.NewReader(ib)); err != nil {
		return err
	}
	opts := oras.PackManifestOptions{
		ConfigDescriptor: &cfg,
		Layers:           layers,
	}
	for _, v := range files {
		l := ocispec.Descriptor{
			MediaType: s.MediaTypeRegistryLayerMetadata(filepath.Base(v.Path())),
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
	logger.C(ctx).Infof("uploaded %s", s.ref)
	return nil
}

func (s *storage) Init(ctx context.Context) error {
	s.lock(ctx)
	defer s.unlock(ctx)
	// if we have a key, we are already initialized
	if s.key != "" {
		return nil
	}
	logger.C(ctx).Infof("initializing %s", s.ref)
	store, err := file.New(s.tmp)
	if err != nil {
		return err
	}
	priv, pub, err := s.repo.GenerateKeypair()
	if err != nil {
		return err
	}
	s.key = priv
	enc, err := aes.Encrypt(s.opts.key, priv)
	if err != nil {
		return err
	}
	var opts oras.PackManifestOptions
	pvn, pbn := s.repo.KeyNames()
	for _, v := range []Artifact{NewFile(pvn, enc), NewFile(pbn, []byte(pub))} {
		l := ocispec.Descriptor{
			MediaType: s.MediaTypeRegistryLayerMetadata(filepath.Base(v.Path())),
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
	logger.C(ctx).Infof("storage initialized %s", s.ref)
	return nil
}

func (s *storage) manifest(ctx context.Context) (m ocispec.Manifest, err error) {
	desc, err := s.rrepo.Resolve(ctx, s.repo.Name())
	if err != nil {
		return m, err
	}
	if v, ok := cache.Get(desc.Digest.String()); ok {
		// reset ttl
		cache.Set(desc.Digest.String(), v, cache.WithTTL(cache.DefaultTTL))
		return v.(ocispec.Manifest), nil
	}
	logger.C(ctx).Infof("retrieve manifest %s", desc.Digest.String())
	b, err := s.rrepo.Manifests().Fetch(ctx, desc)
	if err != nil {
		return m, err
	}
	defer b.Close()

	if err := json.NewDecoder(b).Decode(&m); err != nil {
		return m, err
	}
	if m.ArtifactType != s.ArtefactTypeRegistry() {
		return m, fmt.Errorf("%w: %s", ErrInvalidArtifactType, m.MediaType)
	}
	cache.Set(desc.Digest.String(), m, cache.WithTTL(cache.DefaultTTL))
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
	priv, err := aes.Decrypt(s.opts.key, b)
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

func (s *storage) lock(ctx context.Context) {
	lock.Lock(ctx, s.ref)
}

func (s *storage) unlock(ctx context.Context) {
	lock.Unlock(ctx, s.ref)
}

func (s *storage) artifactName(a Artifact) string {
	return s.name + "/" + strings.Join(slices.Filter([]string{a.Name(), a.Arch(), s.repo.Name()}, func(s string) bool { return s != "" }), "-")
}

func (s *storage) ArtefactTypeRegistry() string {
	return "application/vnd.lk.registry+" + s.repo.Name()
}
func (s *storage) MediaTypeIndexConfig() string {
	return "application/vnd.lk.registry.index.config.v1." + s.repo.Name() + "+json"
}
func (s *storage) MediaTypeArtifactConfig() string {
	return "application/vnd.lk.registry.config.v1." + s.repo.Name() + "+" + s.repo.Codec().Name()
}
func (s *storage) MediaTypeRegistryLayerMetadata(name string) string {
	if ext := strings.TrimPrefix(filepath.Ext(name), "."); ext != "" {
		return "application/vnd.lk.registry.metadata.layer.v1." + s.repo.Name() + "+" + ext
	}
	return "application/vnd.lk.registry.metadata.layer.v1." + s.repo.Name()
}
func (s *storage) MediaTypeArtifactLayer() string {
	return "application/vnd.lk.registry.layer.v1." + s.repo.Name()
}
