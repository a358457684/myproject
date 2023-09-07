package storer

import (
	"context"
	"time"
)

type Storer interface {
	SetToken(ctx context.Context, key, value string, expiration time.Duration) error

	ExistsToken(ctx context.Context, token string) bool

	GetToken(ctx context.Context, key string) (string, error)

	DelToken(ctx context.Context, token string) error

	GetStr(ctx context.Context, key string) (string, error)
}
