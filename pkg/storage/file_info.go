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
	"path/filepath"

	"github.com/opencontainers/go-digest"
)

type info struct {
	version string
	path    string
	size    int64
	digest  digest.Digest
	meta    []byte
}

func (i *info) Name() string {
	return filepath.Base(i.path)
}

func (i *info) Arch() string {
	return ""
}

func (i *info) Version() string {
	return i.version
}

func (i *info) Path() string {
	return i.path
}

func (i *info) Size() int64 {
	return i.size
}

func (i *info) Digest() digest.Digest {
	return i.digest
}

func (i *info) Meta() []byte {
	return i.meta
}
