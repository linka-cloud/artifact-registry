// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package apk

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/opencontainers/go-digest"

	"go.linka.cloud/artifact-registry/pkg/buffer"
	"go.linka.cloud/artifact-registry/pkg/storage"
	"go.linka.cloud/artifact-registry/pkg/validation"
)

var (
	ErrMissingPKGINFOFile = errors.New("PKGINFO file is missing")
	ErrInvalidName        = errors.New("package name is invalid")
	ErrInvalidVersion     = errors.New("package version is invalid")
)

var _ storage.Artifact = (*Package)(nil)

// https://wiki.alpinelinux.org/wiki/Apk_spec

// Package represents an Alpine package
type Package struct {
	PkgName         string          `json:"name"`
	PkgVersion      string          `json:"version"`
	VersionMetadata VersionMetadata `json:"versionMetadata"`
	FileMetadata    FileMetadata    `json:"fileMetadata"`

	PkgSize   int64  `json:"size"`
	PkgDigest string `json:"digest"`
	Branch    string `json:"branch"`
	Repo      string `json:"repo"`
	FilePath  string `json:"filePath"`

	reader io.ReadCloser
}

func (p *Package) Read(b []byte) (n int, err error) {
	if p.reader == nil {
		return 0, io.EOF
	}
	return p.reader.Read(b)
}

func (p *Package) Name() string {
	return p.PkgName
}

func (p *Package) Arch() string {
	return p.FileMetadata.Architecture
}

func (p *Package) Version() string {
	return p.PkgVersion
}

func (p *Package) Path() string {
	return p.FilePath
}

func (p *Package) Size() int64 {
	return p.PkgSize
}

func (p *Package) Digest() digest.Digest {
	return digest.NewDigestFromEncoded(digest.SHA256, p.PkgDigest)
}

func (p *Package) Close() error {
	if p.reader == nil {
		return nil
	}
	return p.reader.Close()
}

// Metadata of an Alpine package
type VersionMetadata struct {
	Maintainer  string `json:"maintainer,omitempty"`
	ProjectURL  string `json:"projectURL,omitempty"`
	Description string `json:"description,omitempty"`
	License     string `json:"license,omitempty"`
}

type FileMetadata struct {
	Checksum     string   `json:"checksum"`
	Packager     string   `json:"packager,omitempty"`
	BuildDate    int64    `json:"buildDate,omitempty"`
	Size         int64    `json:"size,omitempty"`
	Architecture string   `json:"architecture,omitempty"`
	Origin       string   `json:"origin,omitempty"`
	CommitHash   string   `json:"commitHash,omitempty"`
	InstallIf    string   `json:"installIf,omitempty"`
	Provides     []string `json:"provides,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// NewPackage parses the Alpine package file
func NewPackage(r io.Reader, branch, repository string, size int64) (*Package, error) {
	// Alpine packages are concated .tar.gz streams. Usually the first stream contains the package metadata.
	reader, err := buffer.CreateHashedBufferFromReader(r)
	if err != nil {
		return nil, err
	}
	// var buff, tmp bytes.Buffer
	// if _, err := io.Copy(io.MultiWriter(&buff, &tmp), r); err != nil {
	// 	return nil, err
	// }
	br := bufio.NewReader(reader) // needed for gzip Multistream

	h := sha1.New()

	gzr, err := gzip.NewReader(&teeByteReader{br, h})
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	for {
		gzr.Multistream(false)

		tr := tar.NewReader(gzr)
		for {
			hd, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			if hd.Name == ".PKGINFO" {
				p, err := ParsePackageInfo(tr, branch, repository)
				if err != nil {
					return nil, err
				}

				// drain the reader
				for {
					if _, err := tr.Next(); err != nil {
						break
					}
				}

				p.FileMetadata.Checksum = "Q1" + base64.StdEncoding.EncodeToString(h.Sum(nil))
				// p.reader = &buff
				// p.PkgSize = int64(buff.Len())
				_, _, sha256, _ := reader.Sums()
				p.reader = reader
				p.PkgDigest = hex.EncodeToString(sha256)
				p.PkgSize = size
				p.FilePath = fmt.Sprintf("%s/%s/%s/%s-%s.apk", p.Branch, p.Repo, p.FileMetadata.Architecture, p.PkgName, p.PkgVersion)
				_, err = reader.Seek(0, io.SeekStart)
				return p, err
			}
		}

		h = sha1.New()

		err = gzr.Reset(&teeByteReader{br, h})
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return nil, ErrMissingPKGINFOFile
}

// ParsePackageInfo parses a PKGINFO file to retrieve the metadata of an Alpine package
func ParsePackageInfo(r io.Reader, branch, repository string) (*Package, error) {
	p := &Package{
		Branch: branch,
		Repo:   repository,
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			continue
		}

		i := strings.IndexRune(line, '=')
		if i == -1 {
			continue
		}

		key := strings.TrimSpace(line[:i])
		value := strings.TrimSpace(line[i+1:])

		switch key {
		case "pkgname":
			p.PkgName = value
		case "pkgver":
			p.PkgVersion = value
		case "pkgdesc":
			p.VersionMetadata.Description = value
		case "url":
			p.VersionMetadata.ProjectURL = value
		case "builddate":
			n, err := strconv.ParseInt(value, 10, 64)
			if err == nil {
				p.FileMetadata.BuildDate = n
			}
		case "size":
			n, err := strconv.ParseInt(value, 10, 64)
			if err == nil {
				p.FileMetadata.Size = n
			}
		case "arch":
			p.FileMetadata.Architecture = value
		case "origin":
			p.FileMetadata.Origin = value
		case "commit":
			p.FileMetadata.CommitHash = value
		case "maintainer":
			p.VersionMetadata.Maintainer = value
		case "packager":
			p.FileMetadata.Packager = value
		case "license":
			p.VersionMetadata.License = value
		case "install_if":
			p.FileMetadata.InstallIf = value
		case "provides":
			if value != "" {
				p.FileMetadata.Provides = append(p.FileMetadata.Provides, value)
			}
		case "depend":
			if value != "" {
				p.FileMetadata.Dependencies = append(p.FileMetadata.Dependencies, value)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if p.PkgName == "" {
		return nil, ErrInvalidName
	}

	if p.PkgVersion == "" {
		return nil, ErrInvalidVersion
	}

	if !validation.IsValidURL(p.VersionMetadata.ProjectURL) {
		p.VersionMetadata.ProjectURL = ""
	}

	return p, nil
}

// Same as io.TeeReader but implements io.ByteReader
type teeByteReader struct {
	r *bufio.Reader
	w io.Writer
}

func (t *teeByteReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return n, err
}

func (t *teeByteReader) ReadByte() (byte, error) {
	b, err := t.r.ReadByte()
	if err == nil {
		if _, err := t.w.Write([]byte{b}); err != nil {
			return 0, err
		}
	}
	return b, err
}
