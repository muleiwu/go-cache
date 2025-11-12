package test

import (
	"context"
	"testing"
	"time"

	go_cache "github.com/muleiwu/go-cache"
)

// TestNoneExists 测试Exists总是返回false
func TestNoneExists(t *testing.T) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	// 即使设置了键，Exists也应该返回false
	_ = cache.Set(ctx, "test_key", "test_value", 10*time.Minute)

	if cache.Exists(ctx, "test_key") {
		t.Error("None.Exists() 应该总是返回false")
	}

	if cache.Exists(ctx, "any_key") {
		t.Error("None.Exists() 应该总是返回false")
	}
}

// TestNoneGet 测试Get总是返回错误
func TestNoneGet(t *testing.T) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "test_key", &result)
	if err == nil {
		t.Error("None.Get() 应该返回错误")
	}
	if err.Error() != "not implemented" {
		t.Errorf("None.Get() 应该返回 'not implemented' 错误，实际返回: %v", err)
	}
}

// TestNoneSet 测试Set操作成功但不存储数据
func TestNoneSet(t *testing.T) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	// Set应该成功（不返回错误）
	err := cache.Set(ctx, "test_key", "test_value", 10*time.Minute)
	if err != nil {
		t.Errorf("None.Set() 不应该返回错误，实际返回: %v", err)
	}

	// 但是数据不应该被存储
	if cache.Exists(ctx, "test_key") {
		t.Error("None 不应该存储任何数据")
	}
}

// TestNoneGetSet 测试GetSet总是返回错误
func TestNoneGetSet(t *testing.T) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	var result string
	callbackCalled := false

	err := cache.GetSet(ctx, "test_key", 10*time.Minute, &result, func(key string, obj any) error {
		callbackCalled = true
		return nil
	})

	if err == nil {
		t.Error("None.GetSet() 应该返回错误")
	}
	if err.Error() != "not implemented" {
		t.Errorf("None.GetSet() 应该返回 'not implemented' 错误，实际返回: %v", err)
	}

	// 回调函数不应该被调用
	if callbackCalled {
		t.Error("None.GetSet() 的回调函数不应该被调用")
	}
}

// TestNoneDel 测试Del操作成功
func TestNoneDel(t *testing.T) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	// Del应该成功（不返回错误）
	err := cache.Del(ctx, "test_key")
	if err != nil {
		t.Errorf("None.Del() 不应该返回错误，实际返回: %v", err)
	}

	// 删除不存在的键也应该成功
	err = cache.Del(ctx, "non_existent_key")
	if err != nil {
		t.Errorf("None.Del() 不应该返回错误，实际返回: %v", err)
	}
}

// TestNoneExpiresAt 测试ExpiresAt操作成功
func TestNoneExpiresAt(t *testing.T) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	// ExpiresAt应该成功（不返回错误）
	expiresAt := time.Now().Add(10 * time.Minute)
	err := cache.ExpiresAt(ctx, "test_key", expiresAt)
	if err != nil {
		t.Errorf("None.ExpiresAt() 不应该返回错误，实际返回: %v", err)
	}
}

// TestNoneExpiresIn 测试ExpiresIn操作成功
func TestNoneExpiresIn(t *testing.T) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	// ExpiresIn应该成功（不返回错误）
	err := cache.ExpiresIn(ctx, "test_key", 10*time.Minute)
	if err != nil {
		t.Errorf("None.ExpiresIn() 不应该返回错误，实际返回: %v", err)
	}
}

// TestNewCacheNone 测试NewCacheNone别名构造函数
func TestNewCacheNone(t *testing.T) {
	cache := go_cache.NewCacheNone()
	ctx := context.Background()

	// 验证返回的是None类型的实例
	if cache == nil {
		t.Error("NewCacheNone() 不应该返回nil")
	}

	// 验证行为与NewNone()相同
	if cache.Exists(ctx, "test_key") {
		t.Error("NewCacheNone() 创建的实例应该与 NewNone() 行为相同")
	}

	var result string
	err := cache.Get(ctx, "test_key", &result)
	if err == nil {
		t.Error("NewCacheNone() 创建的实例应该与 NewNone() 行为相同")
	}
}

// TestNoneMultipleOperations 测试多个操作的组合
func TestNoneMultipleOperations(t *testing.T) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	// 执行一系列操作
	_ = cache.Set(ctx, "key1", "value1", 10*time.Minute)
	_ = cache.Set(ctx, "key2", 123, 10*time.Minute)
	_ = cache.Set(ctx, "key3", []string{"a", "b", "c"}, 10*time.Minute)

	// 所有键都不应该存在
	if cache.Exists(ctx, "key1") || cache.Exists(ctx, "key2") || cache.Exists(ctx, "key3") {
		t.Error("None 不应该存储任何数据")
	}

	// 尝试获取所有键都应该失败
	var str string
	var num int
	var slice []string

	if err := cache.Get(ctx, "key1", &str); err == nil {
		t.Error("None.Get() 应该总是返回错误")
	}
	if err := cache.Get(ctx, "key2", &num); err == nil {
		t.Error("None.Get() 应该总是返回错误")
	}
	if err := cache.Get(ctx, "key3", &slice); err == nil {
		t.Error("None.Get() 应该总是返回错误")
	}

	// 删除操作应该成功
	_ = cache.Del(ctx, "key1")
	_ = cache.Del(ctx, "key2")
	_ = cache.Del(ctx, "key3")
}

// TestNoneWithDifferentContexts 测试使用不同的context
func TestNoneWithDifferentContexts(t *testing.T) {
	cache := go_cache.NewNone()

	// 使用不同的context
	ctx1 := context.Background()
	ctx2 := context.WithValue(context.Background(), "key", "value")

	// 行为应该相同
	if cache.Exists(ctx1, "test_key") || cache.Exists(ctx2, "test_key") {
		t.Error("None.Exists() 应该总是返回false，不受context影响")
	}

	err1 := cache.Set(ctx1, "test_key", "value", 10*time.Minute)
	err2 := cache.Set(ctx2, "test_key", "value", 10*time.Minute)

	if err1 != nil || err2 != nil {
		t.Error("None.Set() 应该总是成功，不受context影响")
	}
}

// BenchmarkNoneSet 基准测试：Set操作
func BenchmarkNoneSet(b *testing.B) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Set(ctx, "bench_key", "bench_value", 10*time.Minute)
	}
}

// BenchmarkNoneGet 基准测试：Get操作
func BenchmarkNoneGet(b *testing.B) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result string
		_ = cache.Get(ctx, "bench_key", &result)
	}
}

// BenchmarkNoneExists 基准测试：Exists操作
func BenchmarkNoneExists(b *testing.B) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Exists(ctx, "bench_key")
	}
}

// BenchmarkNoneDel 基准测试：Del操作
func BenchmarkNoneDel(b *testing.B) {
	cache := go_cache.NewNone()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Del(ctx, "bench_key")
	}
}
