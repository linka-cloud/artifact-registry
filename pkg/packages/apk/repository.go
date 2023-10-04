// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package apk

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"go.linka.cloud/artifact-registry/pkg/buffer"
	"go.linka.cloud/artifact-registry/pkg/codec"
	rsa2 "go.linka.cloud/artifact-registry/pkg/crypt/rsa"
	"go.linka.cloud/artifact-registry/pkg/slices"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

const (
	RepositoryPublicKey  = "repository.key"
	RepositoryPrivateKey = "private.key"
	IndexFilename        = "APKINDEX.tar.gz"
)

var _ storage.Repository = (*repo)(nil)

type repo struct{}

func (r *repo) Name() string {
	return "apk"
}

func (r *repo) GenerateKeypair() (string, string, error) {
	return rsa2.GenerateKeyPair()
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

// Index (re)builds all repository files for every available distributions, components and architectures
func (r *repo) Index(ctx context.Context, priv string, as ...storage.Artifact) (out []storage.Artifact, err error) {
	pkgs := storage.MustAs[*Package](as)
	branches := slices.Distinct(slices.Map(pkgs, func(v *Package) string {
		return v.Branch
	}))
	for _, branch := range branches {
		pkgs := slices.Filter(pkgs, func(p *Package) bool {
			return p.Branch == branch
		})
		repositories := slices.Distinct(slices.Map(pkgs, func(p *Package) string {
			return p.Repo
		}))
		for _, repository := range repositories {
			pkgs := slices.Filter(pkgs, func(p *Package) bool {
				return p.Repo == repository
			})
			architectures := slices.Distinct(slices.Map(pkgs, func(p *Package) string {
				return p.FileMetadata.Architecture
			}))
			for _, architecture := range architectures {
				a, ok, err := buildPackagesIndex(ctx, branch, repository, architecture, priv, pkgs...)
				if err != nil {
					return nil, fmt.Errorf("failed to build repository files [%s/%s/%s]: %w", branch, repository, architecture, err)
				}
				if ok {
					out = append(out, a)
				}
			}
		}
	}
	return
}

// https://wiki.alpinelinux.org/wiki/Apk_spec#APKINDEX_Format
func buildPackagesIndex(_ context.Context, branch, repository, architecture, priv string, pkgs ...*Package) (storage.Artifact, bool, error) {
	pfs := slices.Filter(pkgs, func(v *Package) bool {
		return v.Branch == branch && v.Repo == repository && v.FileMetadata.Architecture == architecture
	})

	// Delete the package indices if there are no packages
	if len(pfs) == 0 {
		return nil, false, nil
	}

	var buf bytes.Buffer
	for _, pd := range pfs {
		fmt.Fprintf(&buf, "C:%s\n", pd.FileMetadata.Checksum)
		fmt.Fprintf(&buf, "P:%s\n", pd.PkgName)
		fmt.Fprintf(&buf, "V:%s\n", pd.PkgVersion)
		fmt.Fprintf(&buf, "A:%s\n", pd.FileMetadata.Architecture)
		if pd.VersionMetadata.Description != "" {
			fmt.Fprintf(&buf, "T:%s\n", pd.VersionMetadata.Description)
		}
		if pd.VersionMetadata.ProjectURL != "" {
			fmt.Fprintf(&buf, "U:%s\n", pd.VersionMetadata.ProjectURL)
		}
		if pd.VersionMetadata.License != "" {
			fmt.Fprintf(&buf, "L:%s\n", pd.VersionMetadata.License)
		}
		fmt.Fprintf(&buf, "S:%d\n", pd.Size())
		fmt.Fprintf(&buf, "I:%d\n", pd.FileMetadata.Size)
		fmt.Fprintf(&buf, "o:%s\n", pd.FileMetadata.Origin)
		fmt.Fprintf(&buf, "m:%s\n", pd.VersionMetadata.Maintainer)
		fmt.Fprintf(&buf, "t:%d\n", pd.FileMetadata.BuildDate)
		if pd.FileMetadata.CommitHash != "" {
			fmt.Fprintf(&buf, "c:%s\n", pd.FileMetadata.CommitHash)
		}
		if len(pd.FileMetadata.Dependencies) > 0 {
			fmt.Fprintf(&buf, "D:%s\n", strings.Join(pd.FileMetadata.Dependencies, " "))
		}
		if len(pd.FileMetadata.Provides) > 0 {
			fmt.Fprintf(&buf, "p:%s\n", strings.Join(pd.FileMetadata.Provides, " "))
		}
		fmt.Fprint(&buf, "\n")
	}

	unsignedIndexContent, _ := buffer.NewHashedBuffer()
	h := sha1.New()

	if err := writeGzipStream(io.MultiWriter(unsignedIndexContent, h), "APKINDEX", buf.Bytes(), true); err != nil {
		return nil, false, err
	}

	privPem, _ := pem.Decode([]byte(priv))
	if privPem == nil {
		return nil, false, fmt.Errorf("failed to decode private key pem")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		return nil, false, err
	}

	fingerprint, err := rsa2.PublicKeyFingerprint(&privKey.PublicKey)
	if err != nil {
		return nil, false, err
	}

	sign, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA1, h.Sum(nil))
	if err != nil {
		return nil, false, err
	}

	var signedIndexContent bytes.Buffer

	if err := writeGzipStream(
		&signedIndexContent,
		fmt.Sprintf(".SIGN.RSA.lkar@%s.rsa.pub", hex.EncodeToString(fingerprint)),
		sign,
		false,
	); err != nil {
		return nil, false, err
	}

	if _, err := io.Copy(&signedIndexContent, unsignedIndexContent); err != nil {
		return nil, false, err
	}

	return storage.NewFile(filepath.Join(branch, repository, architecture, IndexFilename), signedIndexContent.Bytes()), true, nil
}

func writeGzipStream(w io.Writer, filename string, content []byte, addTarEnd bool) error {
	zw := gzip.NewWriter(w)
	defer zw.Close()

	tw := tar.NewWriter(zw)
	if addTarEnd {
		defer tw.Close()
	}
	hdr := &tar.Header{
		Name: filename,
		Mode: 0o600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(content); err != nil {
		return err
	}
	return nil
}
