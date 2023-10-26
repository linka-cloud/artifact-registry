// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package deb

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"

	"go.linka.cloud/artifact-registry/pkg/buffer"
	"go.linka.cloud/artifact-registry/pkg/codec"
	openpgp2 "go.linka.cloud/artifact-registry/pkg/crypt/openpgp"
	"go.linka.cloud/artifact-registry/pkg/slices"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

const (
	RepositoryPublicKey  = "repository.key"
	RepositoryPrivateKey = "private.key"
)

var _ storage.Repository = (*repo)(nil)

type repo struct{}

func (r *repo) Name() string {
	return "deb"
}

func (r *repo) GenerateKeypair() (string, string, error) {
	return openpgp2.GenerateKeypair("Artifact Registry", "DEB Registry", "")
}

func (r *repo) KeyNames() (string, string) {
	return RepositoryPrivateKey, RepositoryPublicKey
}

func (r *repo) Codec() storage.Codec {
	return codec.Funcs[storage.Artifact]{
		Format: "json",
		EncodeFunc: func(v storage.Artifact) ([]byte, error) {
			return json.Marshal(v)
		},
		DecodeFunc: func(b []byte) (storage.Artifact, error) {
			var v Package
			err := json.Unmarshal(b, &v)
			return &v, err
		},
	}
}

func (r *repo) Index(ctx context.Context, priv string, as ...storage.Artifact) (out []storage.Artifact, err error) {
	pkgs := storage.MustAs[*Package](as)
	distributions := slices.Distinct(slices.Map(pkgs, func(p *Package) string {
		return p.Distribution
	}))
	for _, distribution := range distributions {
		pkgs := slices.Filter(pkgs, func(p *Package) bool {
			return p.Distribution == distribution
		})
		components := slices.Distinct(slices.Map(pkgs, func(p *Package) string {
			return p.Component
		}))
		architectures := slices.Distinct(slices.Map(pkgs, func(p *Package) string {
			return p.Architecture
		}))
		var rs []*storage.File
		for _, component := range components {
			pkgs := slices.Filter(pkgs, func(p *Package) bool {
				return p.Component == component
			})
			for _, architecture := range architectures {
				pkgs := slices.Filter(pkgs, func(p *Package) bool {
					return p.Architecture == architecture
				})
				r, err := buildPackagesIndices(ctx, distribution, component, architecture, pkgs...)
				if err != nil {
					return nil, err
				}
				rs = append(rs, r...)
				out = append(out, storage.AsArtifact(r)...)
			}
		}
		as2, err := buildReleaseFiles(ctx, distribution, components, architectures, priv, rs...)
		if err != nil {
			return nil, err
		}
		out = append(out, as2...)
	}
	return out, nil
}

// https://wiki.debian.org/DebianRepository/Format#A.22Packages.22_Indices
func buildPackagesIndices(_ context.Context, distribution, component, architecture string, pkgs ...*Package) (out []*storage.File, err error) {

	// Delete the package indices if there are no packages
	if len(pkgs) == 0 {
		return nil, nil
	}

	packagesContent := &bytes.Buffer{}

	packagesGzipContent := &bytes.Buffer{}
	gzw := gzip.NewWriter(packagesGzipContent)

	packagesXzContent := &bytes.Buffer{}
	xzw, err := xz.NewWriter(packagesXzContent)
	if err != nil {
		return nil, fmt.Errorf("failed to create xz writer: %w", err)
	}

	w := io.MultiWriter(packagesContent, gzw, xzw)

	addSeparator := false
	for _, v := range pkgs {
		if addSeparator {
			fmt.Fprintln(w)
		}
		addSeparator = true

		fmt.Fprintf(w, "%s\n", strings.TrimSpace(v.Control))

		fmt.Fprintf(w, "Filename: %s\n", v.Path())
		fmt.Fprintf(w, "Size: %d\n", v.PkgSize)
		fmt.Fprintf(w, "MD5sum: %s\n", v.MD5)
		fmt.Fprintf(w, "SHA1: %s\n", v.SHA1)
		fmt.Fprintf(w, "SHA256: %s\n", v.SHA256)
		fmt.Fprintf(w, "SHA512: %s\n", v.SHA512)
	}

	if err := gzw.Close(); err != nil {
		return nil, err
	}
	if err := xzw.Close(); err != nil {
		return nil, err
	}

	for _, v := range []struct {
		name string
		buff *bytes.Buffer
	}{
		{"Packages", packagesContent},
		{"Packages.gz", packagesGzipContent},
		{"Packages.xz", packagesXzContent},
	} {
		out = append(out, storage.NewFile(fmt.Sprintf("dists/%s/%s/binary-%s/%s", distribution, component, architecture, v.name), v.buff.Bytes()))
	}

	return out, nil
}

// https://wiki.debian.org/DebianRepository/Format#A.22Release.22_files
func buildReleaseFiles(_ context.Context, distribution string, components, architectures []string, priv string, files ...*storage.File) (out []storage.Artifact, err error) {
	// Delete the release files if there are no packages
	if len(files) == 0 {
		return nil, nil
	}

	sort.Strings(components)
	sort.Strings(architectures)

	e, err := openpgp2.ParseIdentity(priv)
	if err != nil {
		return nil, err
	}

	inReleaseContent := &bytes.Buffer{}
	if err != nil {
		return nil, err
	}

	sw, err := clearsign.Encode(inReleaseContent, e.PrivateKey, nil)
	if err != nil {
		return nil, err
	}

	releaseContent := &bytes.Buffer{}
	if err != nil {
		return nil, err
	}

	w := io.MultiWriter(sw, releaseContent)

	fmt.Fprintf(w, "Origin: %s\n", "Artifact Registry")
	fmt.Fprintf(w, "Label: %s\n", "Artifact Registry")
	fmt.Fprintf(w, "Suite: %s\n", distribution)
	fmt.Fprintf(w, "Codename: %s\n", distribution)
	fmt.Fprintf(w, "Components: %s\n", strings.Join(components, " "))
	fmt.Fprintf(w, "Architectures: %s\n", strings.Join(architectures, " "))
	fmt.Fprintf(w, "Date: %s\n", time.Now().UTC().Format(time.RFC1123))
	// fmt.Fprintln(w, "Acquire-By-Hash: yes")

	var md5, sha1, sha256, sha512 strings.Builder
	fn := func(v *storage.File) error {
		buff, err := buffer.CreateHashedBufferFromReader(v)
		if err != nil {
			return err
		}
		defer buff.Close()
		data, err := io.ReadAll(buff)
		if err != nil {
			return err
		}
		md5hash, sha1hash, sha256hash, sha512hash := buff.Sums()
		path := strings.TrimPrefix(v.Path(), "dists/"+distribution+"/")
		fmt.Fprintf(&md5, " %x %d %s\n", md5hash, buff.Size(), path)
		fmt.Fprintf(&sha1, " %x %d %s\n", sha1hash, buff.Size(), path)
		fmt.Fprintf(&sha256, " %x %d %s\n", sha256hash, buff.Size(), path)
		fmt.Fprintf(&sha512, " %x %d %s\n", sha512hash, buff.Size(), path)
		// reset the file as we had to read it
		*v = *storage.NewFile(v.Path(), data)
		return nil
	}
	for _, v := range files {
		if err := fn(v); err != nil {
			return nil, fmt.Errorf("%s: %w", v.Path(), err)
		}
	}

	fmt.Fprintln(w, "MD5Sum:")
	fmt.Fprint(w, md5.String())
	fmt.Fprintln(w, "SHA1:")
	fmt.Fprint(w, sha1.String())
	fmt.Fprintln(w, "SHA256:")
	fmt.Fprint(w, sha256.String())
	fmt.Fprintln(w, "SHA512:")
	fmt.Fprint(w, sha512.String())

	sw.Close()

	releaseGpgContent := &bytes.Buffer{}
	if err != nil {
		return nil, err
	}

	if err := openpgp.ArmoredDetachSign(releaseGpgContent, e, bytes.NewReader(releaseContent.Bytes()), nil); err != nil {
		return nil, err
	}

	for _, file := range []struct {
		name string
		data *bytes.Buffer
	}{
		{"Release", releaseContent},
		{"Release.gpg", releaseGpgContent},
		{"InRelease", inReleaseContent},
	} {

		out = append(out, storage.NewFile(filepath.Join("dists", distribution, file.name), file.data.Bytes()))
	}

	return out, nil
}
