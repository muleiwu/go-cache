package go_cache

import (
	"context"
	"errors"
	"time"

	"github.com/muleiwu/gsr"
)

type None struct {
}

func NewCacheNone() *None {
	return NewNone()
}

func NewNone() *None {
	return &None{}
}

func (c *None) Exists(ctx context.Context, key string) bool {
	return false
}

func (c *None) Get(ctx context.Context, key string, obj any) error {
	return errors.New("not implemented")
}

func (c *None) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return nil
}

func (c *None) GetSet(ctx context.Context, key string, ttl time.Duration, obj any, fun gsr.CacheCallback) error {
	return errors.New("not implemented")
}

func (c *None) Del(ctx context.Context, key string) error {
	return nil
}

func (c *None) ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error {
	return nil
}

func (c *None) ExpiresIn(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}
