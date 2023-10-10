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

package main

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"

	"go.linka.cloud/artifact-registry/pkg/packages"
)

func formatSize(v any) string {
	return humanize.Bytes(uint64(v.(int64)))
}

func urlWithType(typ ...string) string {
	base := url()
	if len(typ) == 0 || strings.HasPrefix(registry, typ[0]+".") {
		return base
	}
	return base + "/" + typ[0]
}

func urlHasType() bool {
	return slices.Contains(packages.Providers(), strings.Split(registry, ".")[0])
}

func url() string {
	scheme := "https"
	if plainHTTP {
		scheme = "http"
	}
	return scheme + "://" + registry
}

func repoURL() string {
	if repository == "" {
		return registry
	}
	return registry + "/" + repository
}

func client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}
}

func newProgressReader(r io.Reader, size int64) *prw {
	return &prw{r: r, size: size, closed: make(chan struct{})}
}

type prw struct {
	r      io.Reader
	total  atomic.Int64
	size   int64
	mu     sync.RWMutex
	closed chan struct{}
}

func (p *prw) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	p.total.Add(int64(n))
	return n, err
}

func (p *prw) Progress() int {
	return int(p.total.Load())
}

func (p *prw) Close() error {
	select {
	case <-p.closed:
	default:
		close(p.closed)
	}
	return nil
}

func (p *prw) Run(ctx context.Context) {
	tk := time.NewTicker(time.Second)
	last := 0
	b := p.Progress()
	logrus.Infof("%s / %d%% transfered (%s/s)", humanize.Bytes(uint64(b)), int(float64(b)/float64(p.size)*100), humanize.Bytes(uint64(b-last)))
	last = b
	for {
		select {
		case <-tk.C:
			b := p.Progress()
			logrus.Infof("%s / %d%% transfered (%s/s)", humanize.Bytes(uint64(b)), int(float64(b)/float64(p.size)*100), humanize.Bytes(uint64(b-last)))
			last = b
		case <-p.closed:
			b := p.Progress()
			logrus.Infof("%s / %d%% transfered (%s/s)", humanize.Bytes(uint64(b)), int(float64(b)/float64(p.size)*100), humanize.Bytes(uint64(b-last)))
			return
		case <-ctx.Done():
			return
		}
	}
}
