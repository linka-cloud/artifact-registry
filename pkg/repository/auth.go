package repository

import (
	"context"
)

type Auth interface {
	BasicAuth() (username, password string, ok bool)
}

type authKey struct{}

func ContextWithAuth(ctx context.Context, a Auth) context.Context {
	return context.WithValue(ctx, authKey{}, a)
}

func authFromContext(ctx context.Context) Auth {
	a, _ := ctx.Value(authKey{}).(Auth)
	return a
}
