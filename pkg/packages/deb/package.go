// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package deb

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blakesmith/ar"
	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	"github.com/ulikunitz/xz"

	"go.linka.cloud/artifact-registry/pkg/buffer"
	"go.linka.cloud/artifact-registry/pkg/storage"
	"go.linka.cloud/artifact-registry/pkg/validation"
)

const controlTar = "control.tar"

var (
	ErrMissingControlFile     = errors.New("control file is missing")
	ErrUnsupportedCompression = errors.New("unsupported compression algorithm")
	ErrInvalidName            = errors.New("package name is invalid")
	ErrInvalidVersion         = errors.New("package version is invalid")
	ErrInvalidArchitecture    = errors.New("package architecture is invalid")

	// https://www.debian.org/doc/debian-policy/ch-controlfields.html#source
	namePattern = regexp.MustCompile(`\A[a-z0-9][a-z0-9+-.]+\z`)
	// https://www.debian.org/doc/debian-policy/ch-controlfields.html#version
	versionPattern = regexp.MustCompile(`\A(?:[0-9]:)?[a-zA-Z0-9.+~]+(?:-[a-zA-Z0-9.+-~]+)?\z`)
)

var _ storage.Artifact = (*Package)(nil)

type Package struct {
	PkgName      string    `json:"name"`
	PkgVersion   string    `json:"version"`
	PkgSize      int64     `json:"size"`
	Architecture string    `json:"architecture"`
	Control      string    `json:"control"`
	Metadata     *Metadata `json:"metadata"`

	Component    string `json:"component"`
	Distribution string `json:"distribution"`
	FilePath     string `json:"filePath"`

	MD5    string `json:"md5"`
	SHA1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
	SHA512 string `json:"sha512"`

	reader io.ReadCloser
}

func (p *Package) Read(b []byte) (n int, err error) {
	if p.reader == nil {
		return 0, io.EOF
	}
	return p.reader.Read(b)
}

func (p *Package) Close() error {
	if p.reader == nil {
		return nil
	}
	return p.reader.Close()
}

func (p *Package) Name() string {
	return p.PkgName
}

func (p *Package) Path() string {
	return p.FilePath
}

func (p *Package) Arch() string {
	return p.Architecture
}

func (p *Package) Version() string {
	return p.PkgVersion
}

func (p *Package) Size() int64 {
	return p.PkgSize
}

func (p *Package) Digest() digest.Digest {
	return digest.NewDigestFromEncoded(digest.SHA256, p.SHA256)
}

type Metadata struct {
	Maintainer   string   `json:"maintainer,omitempty"`
	ProjectURL   string   `json:"projectURL,omitempty"`
	Description  string   `json:"description,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// NewPackage parses the Debian package file
// https://manpages.debian.org/bullseye/dpkg-dev/deb.5.en.html
func NewPackage(r io.Reader, distribution, component string, size int64) (*Package, error) {
	reader, err := buffer.CreateHashedBufferFromReader(r)
	if err != nil {
		return nil, err
	}
	pkg, err := parsePackage(reader)
	if err != nil {
		return nil, err
	}
	pkg.Component = component
	pkg.Distribution = distribution
	pkg.FilePath = filepath.Join("pool", pkg.Distribution, pkg.Component, fmt.Sprintf("%s_%s_%s.deb", pkg.PkgName, pkg.PkgVersion, pkg.Architecture))
	pkg.reader = reader
	pkg.PkgSize = size
	md5, sha1, sha256, sha512 := reader.Sums()
	pkg.MD5 = hex.EncodeToString(md5)
	pkg.SHA1 = hex.EncodeToString(sha1)
	pkg.SHA256 = hex.EncodeToString(sha256)
	pkg.SHA512 = hex.EncodeToString(sha512)
	_, err = reader.Seek(0, io.SeekStart)
	return pkg, err
}

func parsePackage(r io.Reader) (*Package, error) {
	arr := ar.NewReader(r)

	for {
		hd, err := arr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(hd.Name, controlTar) {
			var inner io.Reader
			// https://man7.org/linux/man-pages/man5/deb-split.5.html#FORMAT
			// The file names might contain a trailing slash (since dpkg 1.15.6).
			switch strings.TrimSuffix(hd.Name[len(controlTar):], "/") {
			case "":
				inner = arr
			case ".gz":
				gzr, err := gzip.NewReader(arr)
				if err != nil {
					return nil, err
				}
				defer gzr.Close()

				inner = gzr
			case ".xz":
				xzr, err := xz.NewReader(arr)
				if err != nil {
					return nil, err
				}

				inner = xzr
			case ".zst":
				zr, err := zstd.NewReader(arr)
				if err != nil {
					return nil, err
				}
				defer zr.Close()

				inner = zr
			default:
				return nil, ErrUnsupportedCompression
			}

			tr := tar.NewReader(inner)
			for {
				hd, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, err
				}

				if hd.Typeflag != tar.TypeReg {
					continue
				}

				if hd.FileInfo().Name() == "control" {
					return ParseControlFile(tr)
				}
			}
		}
	}

	return nil, ErrMissingControlFile
}

// ParseControlFile parses a Debian control file to retrieve the metadata
func ParseControlFile(r io.Reader) (*Package, error) {
	p := &Package{
		Metadata: &Metadata{},
	}

	key := ""
	var depends strings.Builder
	var control strings.Builder

	s := bufio.NewScanner(io.TeeReader(r, &control))
	for s.Scan() {
		line := s.Text()

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if line[0] == ' ' || line[0] == '\t' {
			switch key {
			case "Description":
				p.Metadata.Description += line
			case "Depends":
				depends.WriteString(trimmed)
			}
		} else {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) < 2 {
				continue
			}

			key = parts[0]
			value := strings.TrimSpace(parts[1])
			switch key {
			case "Package":
				p.PkgName = value
			case "Version":
				p.PkgVersion = value
			case "Architecture":
				p.Architecture = value
			case "Maintainer":
				a, err := mail.ParseAddress(value)
				if err != nil || a.Name == "" {
					p.Metadata.Maintainer = value
				} else {
					p.Metadata.Maintainer = a.Name
				}
			case "Description":
				p.Metadata.Description = value
			case "Depends":
				depends.WriteString(value)
			case "Homepage":
				if validation.IsValidURL(value) {
					p.Metadata.ProjectURL = value
				}
			}
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	if !namePattern.MatchString(p.PkgName) {
		return nil, ErrInvalidName
	}
	if !versionPattern.MatchString(p.PkgVersion) {
		return nil, ErrInvalidVersion
	}
	if p.Architecture == "" {
		return nil, ErrInvalidArchitecture
	}

	dependencies := strings.Split(depends.String(), ",")
	for i := range dependencies {
		dependencies[i] = strings.TrimSpace(dependencies[i])
	}
	p.Metadata.Dependencies = dependencies

	p.Control = strings.TrimSpace(control.String())

	return p, nil
}
