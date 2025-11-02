package go_cache

import (
	"context"
	"time"

	"github.com/muleiwu/go-cache/cache_value"
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	conn *redis.Client
}

func NewRedis(conn *redis.Client) *Redis {
	return &Redis{conn: conn}
}

func (c *Redis) Exists(ctx context.Context, key string) bool {
	exists := c.conn.Exists(ctx, key)

	return exists.Val() != 0
}

func (c *Redis) Get(ctx context.Context, key string, obj any) error {
	cmd := c.conn.Get(ctx, key)

	result, err := cmd.Result()

	if err != nil {
		return err
	}

	err = cache_value.Decode([]byte(result), obj)
	if err != nil {
		return err
	}

	return nil
}

func (c *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	encode, err := cache_value.Encode(value)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = 0
	}
	cmd := c.conn.Set(ctx, key, string(encode), ttl)
	return cmd.Err()
}

func (c *Redis) GetSet(ctx context.Context, key string, ttl time.Duration, obj any, fun gsr.CacheCallback) error {

	err := fun(key, obj)
	if err != nil {
		return err
	}

	return c.Set(ctx, key, obj, ttl)
}

func (c *Redis) Del(ctx context.Context, key string) error {
	return c.conn.Del(ctx, key).Err()
}

func (c *Redis) ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error {
	cmd := c.conn.ExpireAt(ctx, key, expiresAt)
	return cmd.Err()
}

func (c *Redis) ExpiresIn(ctx context.Context, key string, ttl time.Duration) error {
	cmd := c.conn.Expire(ctx, key, ttl)
	return cmd.Err()
}
