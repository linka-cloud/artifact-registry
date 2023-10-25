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
	"runtime"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.linka.cloud/grpc-toolkit/logger"
	"oras.land/oras-go/v2"
)

func copts(name string) oras.CopyOptions {
	var times sync.Map
	return oras.CopyOptions{
		CopyGraphOptions: oras.CopyGraphOptions{
			Concurrency: runtime.NumCPU(),
			PreCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
				times.Store(desc.Digest.String(), time.Now())
				logger.C(ctx).WithFields(
					"digest", desc.Digest.String(),
					"size", humanize.Bytes(uint64(desc.Size)),
					"ref", name,
				).Infof("uploading")
				return nil
			},
			OnCopySkipped: func(ctx context.Context, desc ocispec.Descriptor) error {
				logger.C(ctx).WithFields(
					"digest", desc.Digest.String(),
					"size", humanize.Bytes(uint64(desc.Size)),
					"ref", name,
				).Infof("already exists")
				return nil
			},
			PostCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
				var dur time.Duration
				if v, ok := times.Load(desc.Digest.String()); ok {
					dur = time.Since(v.(time.Time))
				}
				logger.C(ctx).WithFields(
					"digest", desc.Digest.String(),
					"size", humanize.Bytes(uint64(desc.Size)),
					"ref", name,
					"duration", dur,
				).Infof("uploaded")
				return nil
			},
		},
	}
}
