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
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"go.linka.cloud/grpc-toolkit/logger"
)

func formatSize(v any) string {
	return humanize.Bytes(uint64(v.(int64)))
}

func repoURL() string {
	if repository == "" {
		return registry
	}
	return registry + "/" + repository
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
	logger.C(ctx).Infof("%s / %d%% transferred (%s/s)", humanize.Bytes(uint64(b)), int(float64(b)/float64(p.size)*100), humanize.Bytes(uint64(b-last)))
	last = b
	for {
		select {
		case <-tk.C:
			b := p.Progress()
			logger.C(ctx).Infof("%s / %d%% transferred (%s/s)", humanize.Bytes(uint64(b)), int(float64(b)/float64(p.size)*100), humanize.Bytes(uint64(b-last)))
			last = b
		case <-p.closed:
			b := p.Progress()
			logger.C(ctx).Infof("%s / %d%% transferred (%s/s)", humanize.Bytes(uint64(b)), int(float64(b)/float64(p.size)*100), humanize.Bytes(uint64(b-last)))
			return
		case <-ctx.Done():
			return
		}
	}
}
