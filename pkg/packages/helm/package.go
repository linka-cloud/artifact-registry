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

package helm

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/opencontainers/go-digest"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"go.linka.cloud/artifact-registry/pkg/buffer"
	"go.linka.cloud/artifact-registry/pkg/storage"
)

var _ storage.Artifact = (*Package)(nil)

type Package struct {
	*chart.Metadata `json:",inline"`
	PkgDigest       string `json:"digest"`
	PkgSize         int64  `json:"size"`
	FilePath        string `json:"filePath"`

	r io.ReadCloser
}

func (p *Package) Read(b []byte) (n int, err error) {
	if p.r == nil {
		return 0, io.EOF
	}
	return p.r.Read(b)
}

func (p *Package) Close() error {
	if p.r == nil {
		return nil
	}
	return p.r.Close()
}

func (p *Package) Name() string {
	return p.Metadata.Name
}

func (p *Package) Path() string {
	return p.FilePath
}

func (p *Package) Arch() string {
	return "noarch"
}

func (p *Package) Version() string {
	return p.Metadata.Version
}

func (p *Package) Size() int64 {
	return p.PkgSize
}

func (p *Package) Digest() digest.Digest {
	return digest.NewDigestFromEncoded(digest.SHA256, p.PkgDigest)
}

func NewPackage(r io.Reader) (*Package, error) {
	buf, err := buffer.CreateHashedBufferFromReader(r)
	if err != nil {
		return nil, err
	}
	c, err := loader.LoadArchive(buf)
	if err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	_, _, d, _ := buf.Sums()
	if _, err := buf.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return &Package{
		Metadata:  c.Metadata,
		PkgDigest: hex.EncodeToString(d),
		PkgSize:   buf.Size(),
		FilePath:  fmt.Sprintf("%s-%s.tgz", c.Metadata.Name, c.Metadata.Version),
		r:         buf,
	}, nil
}
