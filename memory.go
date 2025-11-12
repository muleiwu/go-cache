package go_cache

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/muleiwu/gsr"
	"github.com/patrickmn/go-cache"
)

type Memory struct {
	cache *cache.Cache
}

func NewMemory(defaultExpiration, cleanupInterval time.Duration) *Memory {
	return &Memory{cache: cache.New(defaultExpiration, cleanupInterval)}
}

func (c *Memory) Exists(ctx context.Context, key string) bool {
	_, b := c.cache.Get(key)
	return b
}

func (c *Memory) Get(ctx context.Context, key string, obj any) error {
	val, b := c.cache.Get(key)
	if !b {
		return errors.New("key not exists")
	}
	return c.assignValue(obj, val)
}

func (c *Memory) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = -1
	}
	c.cache.Set(key, value, ttl)
	return nil
}

func (c *Memory) GetSet(ctx context.Context, key string, ttl time.Duration, obj any, fun gsr.CacheCallback) error {
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

func (c *Memory) Del(ctx context.Context, key string) error {
	c.cache.Delete(key)
	return nil
}

func (c *Memory) ExpiresAt(ctx context.Context, key string, expiresAt time.Time) error {
	// 检查键是否存在
	val, found := c.cache.Get(key)
	if !found {
		return errors.New("key not exists")
	}

	// 计算正确的TTL（过期时间 - 当前时间）
	ttl := time.Until(expiresAt)
	if ttl < 0 {
		// 如果已经过期，删除键
		c.cache.Delete(key)
		return nil
	}

	// 重新设置带新TTL的值
	c.cache.Set(key, val, ttl)

	return nil
}

func (c *Memory) ExpiresIn(ctx context.Context, key string, ttl time.Duration) error {
	// 检查键是否存在
	val, found := c.cache.Get(key)
	if !found {
		return errors.New("key not exists")
	}

	// 重新设置带新TTL的值
	c.cache.Set(key, val, ttl)

	return nil
}

// assignValue 使用反射将值赋给目标对象
func (c *Memory) assignValue(obj any, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("obj cannot be nil")
	}

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		return fmt.Errorf("obj must be a pointer")
	}

	objElem := objValue.Elem()
	if !objElem.CanSet() {
		return fmt.Errorf("obj cannot be set")
	}

	// 如果value是nil，特殊处理
	if value == nil {
		// 如果是指针类型，设置为nil
		if objElem.Kind() == reflect.Ptr ||
			objElem.Kind() == reflect.Slice ||
			objElem.Kind() == reflect.Map ||
			objElem.Kind() == reflect.Chan ||
			objElem.Kind() == reflect.Func ||
			objElem.Kind() == reflect.Interface {
			objElem.Set(reflect.Zero(objElem.Type()))
			return nil
		}
		return fmt.Errorf("cannot assign nil to non-pointer type %s", objElem.Type())
	}

	valueReflect := reflect.ValueOf(value)
	if !valueReflect.IsValid() {
		return fmt.Errorf("value is not valid")
	}

	// 确保类型匹配
	if objElem.Type() != valueReflect.Type() {
		return fmt.Errorf("type mismatch: expected %s, got %s", objElem.Type(), valueReflect.Type())
	}

	objElem.Set(valueReflect)
	return nil
}
