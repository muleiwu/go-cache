package cache_value

import (
	"github.com/muleiwu/go-cache/serializer"
)

var (
	// defaultSerializer 默认序列化器（gob）
	defaultSerializer serializer.Serializer = serializer.NewGob()
)

// SetDefaultSerializer 设置默认序列化器
// 用于向后兼容，如果不使用WithSerializer选项则使用此默认序列化器
func SetDefaultSerializer(s serializer.Serializer) {
	defaultSerializer = s
}

// GetDefaultSerializer 获取默认序列化器
func GetDefaultSerializer() serializer.Serializer {
	return defaultSerializer
}

// Encode 使用默认序列化器序列化缓存值
// 为了向后兼容保留此函数
func Encode(value interface{}) ([]byte, error) {
	return defaultSerializer.Encode(value)
}

// Decode 使用默认序列化器反序列化
// 为了向后兼容保留此函数
func Decode(data []byte, obj any) error {
	return defaultSerializer.Decode(data, obj)
}
