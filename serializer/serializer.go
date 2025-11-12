package serializer

// Serializer 序列化器接口
// 定义了缓存值的编码和解码方法
type Serializer interface {
	// Encode 将值序列化为字节数组
	Encode(value interface{}) ([]byte, error)

	// Decode 将字节数组反序列化为值
	// obj 必须是指针类型
	Decode(data []byte, obj any) error

	// Name 返回序列化器的名称
	Name() string
}
