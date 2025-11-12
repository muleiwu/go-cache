package serializer

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// JsonSerializer JSON序列化器
// 使用Go标准库的encoding/json包
// 优点：人类可读，跨语言支持，易于调试
// 缺点：性能较gob慢，某些类型支持不完整（如复杂指针、interface{}）
type JsonSerializer struct{}

// NewJson 创建JSON序列化器
func NewJson() *JsonSerializer {
	return &JsonSerializer{}
}

// Name 返回序列化器名称
func (j *JsonSerializer) Name() string {
	return "json"
}

// jsonWrapper 包装值以处理nil和类型信息
type jsonWrapper struct {
	IsNil    bool        `json:"is_nil"`
	TypeName string      `json:"type_name,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// Encode 使用JSON序列化缓存值
func (j *JsonSerializer) Encode(value interface{}) ([]byte, error) {
	// 检查是否为nil
	wrapper := jsonWrapper{
		IsNil: value == nil,
		Value: value,
	}

	// 如果value不是nil，检查是否是nil指针/切片/map
	if value != nil {
		valueReflect := reflect.ValueOf(value)
		kind := valueReflect.Kind()

		if (kind == reflect.Ptr || kind == reflect.Slice || kind == reflect.Map) && valueReflect.IsNil() {
			wrapper.IsNil = true
			wrapper.TypeName = valueReflect.Type().String()
			wrapper.Value = nil
		}
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("json encode error: %w", err)
	}

	return data, nil
}

// Decode 使用JSON反序列化
func (j *JsonSerializer) Decode(data []byte, obj any) error {
	if obj == nil {
		return fmt.Errorf("obj cannot be nil")
	}

	// 检查obj必须是指针
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		return fmt.Errorf("obj must be a pointer")
	}

	var wrapper jsonWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("json decode error: %w", err)
	}

	// 如果是nil值
	if wrapper.IsNil {
		objElem := objValue.Elem()
		if !objElem.CanSet() {
			return fmt.Errorf("obj cannot be set")
		}

		// 如果目标类型支持nil，设置为zero value
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

	// 非nil值，直接反序列化到obj
	// 将wrapper.Value重新序列化然后反序列化到obj
	// 这样可以处理类型转换
	valueData, err := json.Marshal(wrapper.Value)
	if err != nil {
		return fmt.Errorf("json re-encode error: %w", err)
	}

	if err := json.Unmarshal(valueData, obj); err != nil {
		return fmt.Errorf("json decode to obj error: %w", err)
	}

	return nil
}
