package packages

import (
	"context"

	"github.com/gorilla/mux"
)

type Provider interface {
	Route(m *mux.Router)
}

type ProviderFactory func(ctx context.Context, backend string, key []byte) (Provider, error)

var providers = map[string]ProviderFactory{}

func Register(name string, factory ProviderFactory) {
	providers[name] = factory
}

func Init(ctx context.Context, r *mux.Router, backend string, key []byte) error {
	for k, v := range providers {
		p, err := v(ctx, backend, key)
		if err != nil {
			return err
		}
		p.Route(r.PathPrefix("/" + k).Subrouter())
	}
	return nil
}
