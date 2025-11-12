package test

import (
	"context"
	"testing"
	"time"

	go_cache "github.com/muleiwu/go-cache"
)

// TestMemorySetAndGet 测试设置和获取缓存
func TestMemorySetAndGet(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	tests := []struct {
		name  string
		key   string
		value interface{}
		ttl   time.Duration
	}{
		{
			name:  "设置字符串",
			key:   "string_key",
			value: "测试字符串",
			ttl:   10 * time.Minute,
		},
		{
			name:  "设置整数",
			key:   "int_key",
			value: 12345,
			ttl:   10 * time.Minute,
		},
		{
			name:  "设置结构体",
			key:   "struct_key",
			value: TestUser{ID: 1, Name: "测试用户", Age: 25},
			ttl:   10 * time.Minute,
		},
		{
			name:  "设置切片",
			key:   "slice_key",
			value: []string{"a", "b", "c"},
			ttl:   10 * time.Minute,
		},
		{
			name:  "设置map",
			key:   "map_key",
			value: map[string]int{"one": 1, "two": 2},
			ttl:   10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置缓存
			err := cache.Set(ctx, tt.key, tt.value, tt.ttl)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			// 获取缓存
			var result interface{}
			switch tt.value.(type) {
			case string:
				var s string
				result = &s
			case int:
				var i int
				result = &i
			case TestUser:
				var u TestUser
				result = &u
			case []string:
				var sl []string
				result = &sl
			case map[string]int:
				var m map[string]int
				result = &m
			}

			err = cache.Get(ctx, tt.key, result)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
		})
	}
}

// TestMemoryExists 测试键是否存在
func TestMemoryExists(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 设置一个键
	err := cache.Set(ctx, "exists_key", "value", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 检查存在的键
	if !cache.Exists(ctx, "exists_key") {
		t.Error("Exists() 应该返回true，但返回了false")
	}

	// 检查不存在的键
	if cache.Exists(ctx, "not_exists_key") {
		t.Error("Exists() 应该返回false，但返回了true")
	}
}

// TestMemoryDel 测试删除缓存
func TestMemoryDel(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 设置一个键
	err := cache.Set(ctx, "del_key", "value", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 确认键存在
	if !cache.Exists(ctx, "del_key") {
		t.Fatal("键应该存在")
	}

	// 删除键
	err = cache.Del(ctx, "del_key")
	if err != nil {
		t.Fatalf("Del() error = %v", err)
	}

	// 确认键已被删除
	if cache.Exists(ctx, "del_key") {
		t.Error("键应该已被删除")
	}
}

// TestMemoryGetSet 测试GetSet方法
func TestMemoryGetSet(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	callCount := 0
	key := "getset_key"

	// 第一次调用，缓存不存在，应该执行回调
	var result1 string
	err := cache.GetSet(ctx, key, 10*time.Minute, &result1, func(k string, obj any) error {
		callCount++
		str := obj.(*string)
		*str = "回调设置的值"
		return nil
	})
	if err != nil {
		t.Fatalf("GetSet() error = %v", err)
	}
	if result1 != "回调设置的值" {
		t.Errorf("GetSet() 值不正确: got %v, want %v", result1, "回调设置的值")
	}
	if callCount != 1 {
		t.Errorf("回调应该被调用1次，实际调用了%d次", callCount)
	}

	// 第二次调用，缓存存在，不应该执行回调
	var result2 string
	err = cache.GetSet(ctx, key, 10*time.Minute, &result2, func(k string, obj any) error {
		callCount++
		return nil
	})
	if err != nil {
		t.Fatalf("GetSet() error = %v", err)
	}
	// 验证缓存命中，回调函数不应被调用
	if callCount != 1 {
		t.Errorf("回调函数不应被调用2次，实际调用了%d次", callCount)
	}
	// 验证返回的值是缓存中的值
	if result2 != "回调设置的值" {
		t.Errorf("应该返回缓存中的值: got %v, want %v", result2, "回调设置的值")
	}
}

// TestMemoryExpiresIn 测试设置相对过期时间
func TestMemoryExpiresIn(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 设置一个键，过期时间为10分钟
	err := cache.Set(ctx, "expire_key", "value", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 验证键存在
	if !cache.Exists(ctx, "expire_key") {
		t.Fatal("键应该存在")
	}

	// 更新过期时间为100毫秒（很短的时间）
	err = cache.ExpiresIn(ctx, "expire_key", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("ExpiresIn() error = %v", err)
	}

	// 等待200毫秒后，键应该已过期
	time.Sleep(200 * time.Millisecond)

	// 检查键是否已过期
	if cache.Exists(ctx, "expire_key") {
		t.Error("键应该已过期")
	}
}

// TestMemoryExpiresAt 测试设置绝对过期时间
func TestMemoryExpiresAt(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 设置一个键
	err := cache.Set(ctx, "expireat_key", "value", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 验证键存在
	if !cache.Exists(ctx, "expireat_key") {
		t.Fatal("键应该存在")
	}

	// 设置过期时间为100毫秒后
	expiresAt := time.Now().Add(100 * time.Millisecond)
	err = cache.ExpiresAt(ctx, "expireat_key", expiresAt)
	if err != nil {
		t.Fatalf("ExpiresAt() error = %v", err)
	}

	// 等待200毫秒后，键应该已过期
	time.Sleep(200 * time.Millisecond)

	// 检查键是否已过期
	if cache.Exists(ctx, "expireat_key") {
		t.Error("键应该已过期")
	}
}

// TestMemoryGetNonExistentKey 测试获取不存在的键
func TestMemoryGetNonExistentKey(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "non_existent_key", &result)
	if err == nil {
		t.Error("Get() 应该返回错误，但没有返回")
	}
}

// TestMemoryTypeMismatch 测试类型不匹配
func TestMemoryTypeMismatch(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 设置一个字符串
	err := cache.Set(ctx, "type_key", "string_value", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 尝试用整数类型获取
	var result int
	err = cache.Get(ctx, "type_key", &result)
	if err == nil {
		t.Error("Get() 应该返回类型不匹配错误，但没有返回")
	}
}

// TestMemoryWithZeroTTL 测试TTL为0或负数的情况
func TestMemoryWithZeroTTL(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 使用0作为TTL
	err := cache.Set(ctx, "zero_ttl_key", "value", 0)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 键应该存在（使用默认过期时间）
	if !cache.Exists(ctx, "zero_ttl_key") {
		t.Error("键应该存在")
	}

	// 使用负数作为TTL
	err = cache.Set(ctx, "negative_ttl_key", "value", -1*time.Second)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 键应该存在
	if !cache.Exists(ctx, "negative_ttl_key") {
		t.Error("键应该存在")
	}
}

// TestMemoryConcurrentAccess 测试并发访问
func TestMemoryConcurrentAccess(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 启动多个goroutine并发写入
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := "concurrent_key"
				err := cache.Set(ctx, key, id*1000+j, 10*time.Minute)
				if err != nil {
					t.Errorf("Set() error = %v", err)
				}

				var result int
				_ = cache.Get(ctx, key, &result)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// BenchmarkMemorySet 基准测试：设置操作
func BenchmarkMemorySet(b *testing.B) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Set(ctx, "bench_key", "bench_value", 10*time.Minute)
	}
}

// BenchmarkMemoryGet 基准测试：获取操作
func BenchmarkMemoryGet(b *testing.B) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 预先设置数据
	_ = cache.Set(ctx, "bench_key", "bench_value", 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result string
		_ = cache.Get(ctx, "bench_key", &result)
	}
}

// BenchmarkMemoryExists 基准测试：检查存在操作
func BenchmarkMemoryExists(b *testing.B) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	// 预先设置数据
	_ = cache.Set(ctx, "bench_key", "bench_value", 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Exists(ctx, "bench_key")
	}
}
