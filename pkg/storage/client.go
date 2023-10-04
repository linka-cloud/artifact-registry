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
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	cache2 "go.linka.cloud/artifact-registry/pkg/cache"
	"go.linka.cloud/artifact-registry/pkg/logger"
	auth2 "go.linka.cloud/artifact-registry/pkg/storage/auth"
)

const (
// plainHTTP = false
// plainHTTP = true
)

var clientCache = cache2.New()

func copts(name string) oras.CopyOptions {
	var times sync.Map
	return oras.CopyOptions{
		CopyGraphOptions: oras.CopyGraphOptions{
			Concurrency: runtime.NumCPU(),
			PreCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
				times.Store(desc.Digest.String(), time.Now())
				logger.C(ctx).WithFields(logrus.Fields{
					"digest": desc.Digest.String(),
					"size":   humanize.Bytes(uint64(desc.Size)),
					"ref":    name,
				}).Infof("uploading")
				return nil
			},
			OnCopySkipped: func(ctx context.Context, desc ocispec.Descriptor) error {
				logger.C(ctx).WithFields(logrus.Fields{
					"digest": desc.Digest.String(),
					"size":   humanize.Bytes(uint64(desc.Size)),
					"ref":    name,
				}).Infof("already exists")
				return nil
			},
			PostCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
				var dur time.Duration
				if v, ok := times.Load(desc.Digest.String()); ok {
					dur = time.Since(v.(time.Time))
				}
				logger.C(ctx).WithFields(logrus.Fields{
					"digest":   desc.Digest.String(),
					"size":     humanize.Bytes(uint64(desc.Size)),
					"ref":      name,
					"duration": dur,
				}).Infof("uploaded")
				return nil
			},
		},
	}
}

func (s *storage) client(ctx context.Context, host string) remote.Client {
	a := auth2.FromContext(ctx)
	if a == nil {
		return http.DefaultClient
	}
	u, p, ok := a.BasicAuth()
	if !ok {
		return http.DefaultClient
	}
	h := sha256.New()
	h.Write([]byte(u))
	h.Write([]byte(p))
	h.Write([]byte(host))
	key := fmt.Sprintf("%x", h.Sum(nil))
	if v, ok := clientCache.Get(key); ok {
		clientCache.Set(key, v)
		return v.(remote.Client)
	}
	c := &auth.Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: s.opts.insecure,
				},
			},
		},
		// expectedHostAddress is of form ipaddr:port
		Credential: auth.StaticCredential(host, auth.Credential{
			Username: u,
			Password: p,
		}),
		// Cache caches credentials for accessing the remote registry.
		Cache: auth.NewCache(),
	}
	clientCache.Set(key, c)
	return c
}
