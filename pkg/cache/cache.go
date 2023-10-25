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

package cache

import (
	"context"
	"sync"
	"time"
)

type item struct {
	value any
	exp   time.Time
}

type Option func(*item)

func WithTTL(d time.Duration) Option {
	return func(i *item) {
		i.exp = time.Now().Add(d)
	}
}

type Cache interface {
	Set(key string, value any, opts ...Option)
	Get(key string) (any, bool)
	Close() error
}

type cache struct {
	m      sync.Map
	cancel context.CancelFunc
}

func (c *cache) Set(key string, value any, opts ...Option) {
	i := &item{value: value}
	for _, o := range opts {
		o(i)
	}
	c.m.Store(key, i)
}

func (c *cache) Get(key string) (any, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return nil, false
	}
	i := v.(*item)
	if !i.exp.IsZero() && i.exp.Before(time.Now()) {
		c.m.Delete(key)
		return nil, false
	}
	return i.value, true
}

func (c *cache) Close() error {
	c.cancel()
	return nil
}

func (c *cache) run(ctx context.Context) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case now := <-t.C:
			c.m.Range(func(key any, value any) bool {
				i := value.(*item)
				if !i.exp.IsZero() && i.exp.Before(now) {
					c.m.Delete(key)
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}

func New() Cache {
	ctx, cancel := context.WithCancel(context.Background())

	c := &cache{cancel: cancel}
	go c.run(ctx)
	return c
}
