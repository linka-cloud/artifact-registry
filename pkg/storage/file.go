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
	"io"
	"path/filepath"

	"github.com/opencontainers/go-digest"
)

var _ Artifact = (*File)(nil)

type File struct {
	name string
	data []byte
	r    io.Reader
}

func NewFile(name string, data []byte) *File {
	return &File{
		name: name,
		data: data,
		r:    bytes.NewReader(data),
	}
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.r.Read(p)
}

func (f *File) Name() string {
	return filepath.Base(f.name)
}

func (f *File) Arch() string {
	return ""
}

func (f *File) Version() string {
	return ""
}

func (f *File) Path() string {
	return f.name
}

func (f *File) Size() int64 {
	return int64(len(f.data))
}

func (f *File) Digest() digest.Digest {
	return digest.FromBytes(f.data)
}

func (f *File) Close() error {
	return nil
}
