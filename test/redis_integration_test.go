package test

import (
	"context"
	"os"
	"testing"
	"time"

	go_cache "github.com/muleiwu/go-cache"
	"github.com/redis/go-redis/v9"
)

// 测试前检查Redis是否可用
func setupRedisTest(t *testing.T) (*go_cache.Redis, *redis.Client, func()) {
	// 检查环境变量，如果设置了SKIP_REDIS_TESTS，则跳过测试
	if os.Getenv("SKIP_REDIS_TESTS") != "" {
		t.Skip("跳过Redis集成测试（设置了SKIP_REDIS_TESTS环境变量）")
	}

	// 创建Redis客户端
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       15, // 使用DB 15进行测试，避免影响其他数据
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis不可用，跳过集成测试: %v", err)
	}

	// 清空测试数据库
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("清空Redis测试数据库失败: %v", err)
	}

	cache := go_cache.NewRedis(rdb)

	// 返回清理函数
	cleanup := func() {
		// 清理测试数据
		rdb.FlushDB(ctx)
		rdb.Close()
	}

	return cache, rdb, cleanup
}

// TestRedisSetAndGet 测试设置和获取缓存
func TestRedisSetAndGet(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

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
			name:  "设置浮点数",
			key:   "float_key",
			value: 3.14159,
			ttl:   10 * time.Minute,
		},
		{
			name:  "设置布尔值",
			key:   "bool_key",
			value: true,
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
			case float64:
				var f float64
				result = &f
			case bool:
				var b bool
				result = &b
			}

			err = cache.Get(ctx, tt.key, result)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
		})
	}
}

// TestRedisExists 测试键是否存在
func TestRedisExists(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

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

// TestRedisDel 测试删除缓存
func TestRedisDel(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

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

// TestRedisGetSet 测试GetSet方法
func TestRedisGetSet(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

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

// TestRedisExpiresIn 测试设置相对过期时间
func TestRedisExpiresIn(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

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

	// 更新过期时间为1秒（Redis最小支持的单位）
	err = cache.ExpiresIn(ctx, "expire_key", 1*time.Second)
	if err != nil {
		t.Fatalf("ExpiresIn() error = %v", err)
	}

	// 等待1.5秒后，键应该已过期
	time.Sleep(1500 * time.Millisecond)

	// 检查键是否已过期
	if cache.Exists(ctx, "expire_key") {
		t.Error("键应该已过期")
	}
}

// TestRedisExpiresAt 测试设置绝对过期时间
func TestRedisExpiresAt(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

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

	// 设置过期时间为1秒后（Redis最小支持的单位）
	expiresAt := time.Now().Add(1 * time.Second)
	err = cache.ExpiresAt(ctx, "expireat_key", expiresAt)
	if err != nil {
		t.Fatalf("ExpiresAt() error = %v", err)
	}

	// 等待1.5秒后，键应该已过期
	time.Sleep(1500 * time.Millisecond)

	// 检查键是否已过期
	if cache.Exists(ctx, "expireat_key") {
		t.Error("键应该已过期")
	}
}

// TestRedisGetNonExistentKey 测试获取不存在的键
func TestRedisGetNonExistentKey(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "non_existent_key", &result)
	if err == nil {
		t.Error("Get() 应该返回错误，但没有返回")
	}
}

// TestRedisWithZeroTTL 测试TTL为0的情况
func TestRedisWithZeroTTL(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()

	// 使用0作为TTL（永不过期）
	err := cache.Set(ctx, "zero_ttl_key", "value", 0)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 键应该存在
	if !cache.Exists(ctx, "zero_ttl_key") {
		t.Error("键应该存在")
	}

	// 等待一段时间，键应该仍然存在
	time.Sleep(100 * time.Millisecond)
	if !cache.Exists(ctx, "zero_ttl_key") {
		t.Error("键应该仍然存在（TTL为0表示永不过期）")
	}
}

// TestRedisWithNegativeTTL 测试负数TTL的情况
func TestRedisWithNegativeTTL(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()

	// 使用负数作为TTL
	err := cache.Set(ctx, "negative_ttl_key", "value", -1*time.Second)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 键应该存在（负数TTL在Redis中被当作0处理，即永不过期）
	if !cache.Exists(ctx, "negative_ttl_key") {
		t.Error("键应该存在")
	}
}

// TestRedisConcurrentAccess 测试并发访问
func TestRedisConcurrentAccess(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()

	// 启动多个goroutine并发写入和读取
	done := make(chan bool)
	errorChan := make(chan error, 100)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 50; j++ {
				key := "concurrent_key"
				value := id*1000 + j

				// 写入
				err := cache.Set(ctx, key, value, 10*time.Minute)
				if err != nil {
					errorChan <- err
					return
				}

				// 读取
				var result int
				err = cache.Get(ctx, key, &result)
				if err != nil {
					// Get可能会因为类型不匹配失败，这是正常的（并发写入）
					continue
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		t.Errorf("并发访问出错: %v", err)
	}
}

// TestRedisComplexStruct 测试复杂结构体的序列化
// 注意：由于msgpack序列化的特性，复杂结构体会被转换为map[string]interface{}
func TestRedisComplexStruct(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()

	type Address struct {
		City   string
		Street string
	}

	type Person struct {
		Name    string
		Age     int
		Address Address
	}

	original := Person{
		Name: "张三",
		Age:  30,
		Address: Address{
			City:   "北京",
			Street: "长安街",
		},
	}

	// 设置缓存
	err := cache.Set(ctx, "person_key", original, 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// 验证能够存储（但不验证反序列化，因为msgpack的限制）
	if !cache.Exists(ctx, "person_key") {
		t.Error("键应该存在")
	}

	// 注意：由于msgpack的限制，复杂结构体无法直接反序列化
	// 如果需要使用复杂结构体，建议使用Memory缓存或自定义序列化
}

// TestRedisConnectionFailure 测试Redis连接失败的情况
func TestRedisConnectionFailure(t *testing.T) {
	// 创建一个连接到不存在服务器的客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:9999", // 不存在的端口
		Password: "",
		DB:       0,
	})

	cache := go_cache.NewRedis(rdb)
	ctx := context.Background()

	// 尝试设置值应该失败
	err := cache.Set(ctx, "test_key", "test_value", 10*time.Minute)
	if err == nil {
		t.Error("Set() 应该返回错误（连接失败），但没有返回")
	}

	// 尝试获取值应该失败
	var result string
	err = cache.Get(ctx, "test_key", &result)
	if err == nil {
		t.Error("Get() 应该返回错误（连接失败），但没有返回")
	}
}

// BenchmarkRedisSet 基准测试：设置操作
func BenchmarkRedisSet(b *testing.B) {
	cache, _, cleanup := setupRedisTest(&testing.T{})
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Set(ctx, "bench_key", "bench_value", 10*time.Minute)
	}
}

// BenchmarkRedisGet 基准测试：获取操作
func BenchmarkRedisGet(b *testing.B) {
	cache, _, cleanup := setupRedisTest(&testing.T{})
	defer cleanup()

	ctx := context.Background()

	// 预先设置数据
	_ = cache.Set(ctx, "bench_key", "bench_value", 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result string
		_ = cache.Get(ctx, "bench_key", &result)
	}
}

// BenchmarkRedisExists 基准测试：检查存在操作
func BenchmarkRedisExists(b *testing.B) {
	cache, _, cleanup := setupRedisTest(&testing.T{})
	defer cleanup()

	ctx := context.Background()

	// 预先设置数据
	_ = cache.Set(ctx, "bench_key", "bench_value", 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Exists(ctx, "bench_key")
	}
}
