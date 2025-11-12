package test

import (
	"context"
	"testing"
	"time"

	"github.com/muleiwu/go-cache"
	"github.com/muleiwu/go-cache/serializer"
	"github.com/redis/go-redis/v9"
)

// TestUser 测试用户结构体
type TestSerializerUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// TestGobSerializer 测试Gob序列化器
func TestGobSerializer(t *testing.T) {
	gobSer := serializer.NewGob()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "字符串",
			value:   "test_string",
			wantErr: false,
		},
		{
			name:    "整数",
			value:   123,
			wantErr: false,
		},
		{
			name:    "结构体",
			value:   TestSerializerUser{ID: 1, Name: "张三", Age: 25},
			wantErr: false,
		},
		{
			name:    "切片",
			value:   []int{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "map",
			value:   map[string]int{"a": 1, "b": 2},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 编码
			data, err := gobSer.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// 解码
			var result interface{}
			switch tt.value.(type) {
			case string:
				var s string
				result = &s
			case int:
				var i int
				result = &i
			case TestSerializerUser:
				var u TestSerializerUser
				result = &u
			case []int:
				var sl []int
				result = &sl
			case map[string]int:
				var m map[string]int
				result = &m
			}

			err = gobSer.Decode(data, result)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
			}
		})
	}
}

// TestJsonSerializer 测试JSON序列化器
func TestJsonSerializer(t *testing.T) {
	jsonSer := serializer.NewJson()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "字符串",
			value:   "test_string",
			wantErr: false,
		},
		{
			name:    "整数",
			value:   123,
			wantErr: false,
		},
		{
			name:    "结构体",
			value:   TestSerializerUser{ID: 1, Name: "张三", Age: 25},
			wantErr: false,
		},
		{
			name:    "指针",
			value:   &TestSerializerUser{ID: 2, Name: "李四", Age: 30},
			wantErr: false,
		},
		{
			name:    "nil指针",
			value:   (*TestSerializerUser)(nil),
			wantErr: false,
		},
		{
			name:    "切片",
			value:   []int{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "map",
			value:   map[string]int{"a": 1, "b": 2},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 编码
			data, err := jsonSer.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// 解码
			var result interface{}
			switch tt.value.(type) {
			case string:
				var s string
				result = &s
			case int:
				var i int
				result = &i
			case TestSerializerUser:
				var u TestSerializerUser
				result = &u
			case *TestSerializerUser:
				var u *TestSerializerUser
				result = &u
			case []int:
				var sl []int
				result = &sl
			case map[string]int:
				var m map[string]int
				result = &m
			}

			err = jsonSer.Decode(data, result)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
			}
		})
	}
}

// TestRedisWithJsonSerializer 测试Redis使用JSON序列化器
func TestRedisWithJsonSerializer(t *testing.T) {
	// 尝试连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // 使用测试专用DB
	})
	defer rdb.Close()

	ctx := context.Background()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
		return
	}

	// 清空测试DB
	defer rdb.FlushDB(ctx)

	// 创建使用JSON序列化器的Redis缓存
	cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewJson()))

	// 测试基本操作
	t.Run("Set和Get字符串", func(t *testing.T) {
		key := "test:json:string"
		value := "Hello JSON"

		err := cache.Set(ctx, key, value, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result string
		err = cache.Get(ctx, key, &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != value {
			t.Errorf("Get() = %v, want %v", result, value)
		}
	})

	t.Run("Set和Get结构体", func(t *testing.T) {
		key := "test:json:struct"
		value := TestSerializerUser{ID: 100, Name: "JSON User", Age: 28}

		err := cache.Set(ctx, key, value, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result TestSerializerUser
		err = cache.Get(ctx, key, &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != value {
			t.Errorf("Get() = %+v, want %+v", result, value)
		}
	})

	t.Run("Set和Get nil值", func(t *testing.T) {
		key := "test:json:nil"
		var value *TestSerializerUser // nil指针

		err := cache.Set(ctx, key, value, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result *TestSerializerUser
		err = cache.Get(ctx, key, &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != nil {
			t.Errorf("Get() = %v, want nil", result)
		}
	})
}

// TestRedisWithGobSerializer 测试Redis使用Gob序列化器（默认）
func TestRedisWithGobSerializer(t *testing.T) {
	// 尝试连接Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})
	defer rdb.Close()

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
		return
	}

	defer rdb.FlushDB(ctx)

	// 创建使用Gob序列化器的Redis缓存（显式指定）
	cache := go_cache.NewRedis(rdb, go_cache.WithRedisSerializer(serializer.NewGob()))

	t.Run("Set和Get复杂结构体", func(t *testing.T) {
		key := "test:gob:struct"
		value := TestSerializerUser{ID: 200, Name: "Gob User", Age: 35}

		err := cache.Set(ctx, key, value, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		var result TestSerializerUser
		err = cache.Get(ctx, key, &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if result != value {
			t.Errorf("Get() = %+v, want %+v", result, value)
		}
	})
}

// BenchmarkGobSerializer 基准测试Gob序列化器
func BenchmarkGobSerializer(b *testing.B) {
	gobSer := serializer.NewGob()
	value := TestSerializerUser{ID: 1, Name: "Benchmark User", Age: 30}

	b.Run("Encode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gobSer.Encode(value)
		}
	})

	data, _ := gobSer.Encode(value)
	b.Run("Decode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result TestSerializerUser
			_ = gobSer.Decode(data, &result)
		}
	})
}

// BenchmarkJsonSerializer 基准测试JSON序列化器
func BenchmarkJsonSerializer(b *testing.B) {
	jsonSer := serializer.NewJson()
	value := TestSerializerUser{ID: 1, Name: "Benchmark User", Age: 30}

	b.Run("Encode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = jsonSer.Encode(value)
		}
	})

	data, _ := jsonSer.Encode(value)
	b.Run("Decode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result TestSerializerUser
			_ = jsonSer.Decode(data, &result)
		}
	})
}
