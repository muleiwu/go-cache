package test

import (
	"testing"

	"github.com/muleiwu/go-cache/cache_value"
)

// 测试结构体
type TestUser struct {
	ID   int
	Name string
	Age  int
}

// TestEncode 测试编码功能
func TestEncode(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "编码字符串",
			value:   "test_string",
			wantErr: false,
		},
		{
			name:    "编码整数",
			value:   123,
			wantErr: false,
		},
		{
			name:    "编码结构体",
			value:   TestUser{ID: 1, Name: "张三", Age: 25},
			wantErr: false,
		},
		{
			name:    "编码指针",
			value:   &TestUser{ID: 2, Name: "李四", Age: 30},
			wantErr: false,
		},
		{
			name:    "编码切片",
			value:   []int{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "编码map",
			value:   map[string]int{"a": 1, "b": 2},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := cache_value.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) == 0 {
				t.Error("Encode() 返回空数据")
			}
		})
	}
}

// TestDecode 测试解码功能
func TestDecode(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		createObj func() interface{}
		wantErr   bool
	}{
		{
			name:  "解码字符串",
			value: "test_string",
			createObj: func() interface{} {
				var s string
				return &s
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 先编码
			data, err := cache_value.Encode(tt.value)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// 再解码
			obj := tt.createObj()
			err = cache_value.Decode(data, obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDecodeTypeMismatch 测试类型不匹配的情况
func TestDecodeTypeMismatch(t *testing.T) {
	// 编码一个字符串
	data, err := cache_value.Encode("test_string")
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// 尝试解码为整数（应该失败）
	var i int
	err = cache_value.Decode(data, &i)
	if err == nil {
		t.Error("Decode() 期望返回类型不匹配错误，但没有返回错误")
	}
}

// TestDecodeWithNilObj 测试传入nil对象
func TestDecodeWithNilObj(t *testing.T) {
	data, err := cache_value.Encode("test")
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	err = cache_value.Decode(data, nil)
	if err == nil {
		t.Error("Decode() 期望返回nil对象错误，但没有返回错误")
	}
}

// TestDecodeWithNonPointer 测试传入非指针对象
func TestDecodeWithNonPointer(t *testing.T) {
	data, err := cache_value.Encode("test")
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var s string
	err = cache_value.Decode(data, s)
	if err == nil {
		t.Error("Decode() 期望返回非指针错误，但没有返回错误")
	}
}

// cache_value 包使用可插拔的序列化系统
// 默认的 Gob 序列化器完整支持所有 Go 类型，包括复杂结构体、切片、map 等
// Redis 缓存也支持 JSON 序列化器作为替代选项
