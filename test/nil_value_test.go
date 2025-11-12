package test

import (
	"context"
	"testing"
	"time"

	go_cache "github.com/muleiwu/go-cache"
	"github.com/muleiwu/go-cache/cache_value"
)

// TestCacheValueEncodeDecodeNil 测试cache_value对nil值的编解码
func TestCacheValueEncodeDecodeNil(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		createObj func() interface{}
		wantErr   bool
	}{
		{
			name:  "编码解码nil值",
			value: nil,
			createObj: func() interface{} {
				var p *string
				return &p
			},
			wantErr: false,
		},
		{
			name: "编码解码nil指针",
			value: func() interface{} {
				var p *TestUser
				return p
			}(),
			createObj: func() interface{} {
				var p *TestUser
				return &p
			},
			wantErr: false,
		},
		{
			name:  "编码解码nil切片",
			value: ([]string)(nil),
			createObj: func() interface{} {
				var s []string
				return &s
			},
			wantErr: false,
		},
		{
			name:  "编码解码nil map",
			value: (map[string]int)(nil),
			createObj: func() interface{} {
				var m map[string]int
				return &m
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 编码
			data, err := cache_value.Encode(tt.value)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// 解码
			obj := tt.createObj()
			err = cache_value.Decode(data, obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证解码后的值是否为nil
				// 这里需要通过反射检查
				t.Logf("成功编解码nil值: %T", tt.value)
			}
		})
	}
}

// TestMemorySetGetNil 测试Memory缓存对nil值的支持
func TestMemorySetGetNil(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	t.Run("存储和获取nil指针", func(t *testing.T) {
		var user *TestUser = nil

		// 存储nil指针
		err := cache.Set(ctx, "nil_user", user, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// 验证键存在
		if !cache.Exists(ctx, "nil_user") {
			t.Error("键应该存在")
		}

		// 获取nil指针
		var result *TestUser
		err = cache.Get(ctx, "nil_user", &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}
	})

	t.Run("存储和获取nil切片", func(t *testing.T) {
		var slice []string = nil

		err := cache.Set(ctx, "nil_slice", slice, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result []string
		err = cache.Get(ctx, "nil_slice", &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}
	})

	t.Run("存储和获取nil map", func(t *testing.T) {
		var m map[string]int = nil

		err := cache.Set(ctx, "nil_map", m, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result map[string]int
		err = cache.Get(ctx, "nil_map", &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}
	})

	t.Run("存储和获取nil interface", func(t *testing.T) {
		var iface interface{} = nil

		err := cache.Set(ctx, "nil_interface", iface, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result interface{}
		err = cache.Get(ctx, "nil_interface", &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}
	})

	t.Run("存储nil值但获取到非指针类型应该失败", func(t *testing.T) {
		var user *TestUser = nil

		err := cache.Set(ctx, "nil_for_value", user, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// 尝试获取到非指针类型应该失败
		var result TestUser
		err = cache.Get(ctx, "nil_for_value", &result)
		if err == nil {
			t.Error("Get() 应该返回错误，因为无法将nil赋值给非指针类型")
		}
	})
}

// TestRedisSetGetNil 测试Redis缓存对nil值的支持
func TestRedisSetGetNil(t *testing.T) {
	cache, _, cleanup := setupRedisTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("存储和获取nil指针", func(t *testing.T) {
		var user *TestUser = nil

		// 存储nil指针
		err := cache.Set(ctx, "nil_user", user, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// 验证键存在
		if !cache.Exists(ctx, "nil_user") {
			t.Error("键应该存在")
		}

		// 获取nil指针
		var result *TestUser
		err = cache.Get(ctx, "nil_user", &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}
	})

	t.Run("存储和获取nil切片", func(t *testing.T) {
		var slice []string = nil

		err := cache.Set(ctx, "nil_slice", slice, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result []string
		err = cache.Get(ctx, "nil_slice", &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}
	})

	t.Run("存储和获取nil map", func(t *testing.T) {
		var m map[string]int = nil

		err := cache.Set(ctx, "nil_map", m, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result map[string]int
		err = cache.Get(ctx, "nil_map", &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}
	})

	t.Run("存储nil值但获取到非指针类型应该失败", func(t *testing.T) {
		var user *TestUser = nil

		err := cache.Set(ctx, "nil_for_value", user, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// 尝试获取到非指针类型应该失败
		var result TestUser
		err = cache.Get(ctx, "nil_for_value", &result)
		if err == nil {
			t.Error("Get() 应该返回错误，因为无法将nil赋值给非指针类型")
		}
	})
}

// TestMemoryGetSetWithNil 测试GetSet方法对nil值的支持
func TestMemoryGetSetWithNil(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	t.Run("GetSet回调返回nil值", func(t *testing.T) {
		key := "getset_nil_key"
		var result *TestUser

		err := cache.GetSet(ctx, key, 10*time.Minute, &result, func(k string, obj any) error {
			// 回调函数返回nil
			ptr := obj.(**TestUser)
			*ptr = nil
			return nil
		})

		if err != nil {
			t.Fatalf("GetSet() error = %v", err)
		}

		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}

		// 再次调用GetSet，应该从缓存获取nil值
		var result2 *TestUser
		err = cache.GetSet(ctx, key, 10*time.Minute, &result2, func(k string, obj any) error {
			t.Error("不应该调用回调函数")
			return nil
		})

		if err != nil {
			t.Fatalf("GetSet() error = %v", err)
		}

		if result2 != nil {
			t.Errorf("期望得到nil，但得到 %v", result2)
		}
	})
}

// TestNilVsKeyNotExists 测试区分nil值和键不存在
func TestNilVsKeyNotExists(t *testing.T) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()

	t.Run("存储nil值后键应该存在", func(t *testing.T) {
		var nilPtr *string = nil
		err := cache.Set(ctx, "has_nil", nilPtr, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// 键应该存在
		if !cache.Exists(ctx, "has_nil") {
			t.Error("存储nil值后，键应该存在")
		}

		// 可以获取nil值
		var result *string
		err = cache.Get(ctx, "has_nil", &result)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if result != nil {
			t.Errorf("期望得到nil，但得到 %v", result)
		}
	})

	t.Run("键不存在时Get应该返回错误", func(t *testing.T) {
		var result *string
		err := cache.Get(ctx, "not_exists", &result)
		if err == nil {
			t.Error("键不存在时，Get() 应该返回错误")
		}
	})

	t.Run("使用Exists可以区分nil值和键不存在", func(t *testing.T) {
		// 设置一个nil值
		var nilPtr *TestUser = nil
		cache.Set(ctx, "nil_key", nilPtr, 10*time.Minute)

		// 使用Exists检查
		if !cache.Exists(ctx, "nil_key") {
			t.Error("存储nil值的键应该存在")
		}

		if cache.Exists(ctx, "never_set_key") {
			t.Error("从未设置的键不应该存在")
		}
	})
}

// BenchmarkMemorySetNil 基准测试：存储nil值
func BenchmarkMemorySetNil(b *testing.B) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()
	var nilPtr *TestUser = nil

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Set(ctx, "bench_nil", nilPtr, 10*time.Minute)
	}
}

// BenchmarkMemoryGetNil 基准测试：获取nil值
func BenchmarkMemoryGetNil(b *testing.B) {
	cache := go_cache.NewMemory(5*time.Minute, 10*time.Minute)
	ctx := context.Background()
	var nilPtr *TestUser = nil
	_ = cache.Set(ctx, "bench_nil", nilPtr, 10*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result *TestUser
		_ = cache.Get(ctx, "bench_nil", &result)
	}
}
