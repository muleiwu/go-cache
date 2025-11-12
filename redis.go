package go_cache

import (
	"context"
	"reflect"
	"time"

	"github.com/muleiwu/go-cache/cache_value"
	"github.com/muleiwu/go-cache/serializer"
	"github.com/muleiwu/gsr"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	conn       *redis.Client
	serializer serializer.Serializer
}

// RedisOption Redis缓存选项
type RedisOption func(*Redis)

// WithRedisSerializer 设置Redis缓存的序列化器
func WithRedisSerializer(s serializer.Serializer) RedisOption {
	return func(r *Redis) {
		r.serializer = s
	}
}

// NewRedis 创建Redis缓存实例
// 默认使用gob序列化器
func NewRedis(conn *redis.Client, opts ...RedisOption) *Redis {
	r := &Redis{
		conn:       conn,
		serializer: cache_value.GetDefaultSerializer(), // 默认使用gob
	}

	// 应用选项
	for _, opt := range opts {
		opt(r)
	}

	return r
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

	err = c.serializer.Decode([]byte(result), obj)
	if err != nil {
		return err
	}

	return nil
}

func (c *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	encode, err := c.serializer.Encode(value)
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
	// 先尝试从缓存获取
	err := c.Get(ctx, key, obj)
	if err == nil {
		// 缓存命中，直接返回
		return nil
	}

	// 缓存未命中，调用回调函数
	err = fun(key, obj)
	if err != nil {
		return err
	}

	// 获取obj指向的实际值并存入缓存
	// obj是一个指针，我们需要存储它指向的值
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}
	return c.Set(ctx, key, objValue.Interface(), ttl)
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
