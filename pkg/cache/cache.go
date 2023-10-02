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

func Get[T any](c Cache, key string) (z T, ok bool) {
	v, ok := c.Get(key)
	if !ok {
		return z, false
	}
	z, ok = v.(T)
	return
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
