package cache_value

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"sync"
)

var (
	// 记录已注册的类型
	registeredTypes sync.Map
)

// nilValueMarker 用于标记nil值（nil指针/nil切片/nil map）
type nilValueMarker struct {
	TypeName string
}

func init() {
	// 注册nilValueMarker类型
	gob.Register(&nilValueMarker{})
}

// registerTypeIfNeeded 安全地注册类型
func registerTypeIfNeeded(value interface{}) {
	if value == nil {
		return
	}

	valueType := reflect.TypeOf(value)
	typeName := valueType.String()

	// 已注册过则跳过
	if _, loaded := registeredTypes.LoadOrStore(typeName, true); loaded {
		return
	}

	// 使用defer recover捕获panic（重复注册会导致panic）
	defer func() {
		if r := recover(); r != nil {
			// 忽略重复注册的错误
		}
	}()

	// 注册类型
	gob.Register(value)
}

// Encode 使用gob序列化缓存值
func Encode(value interface{}) ([]byte, error) {
	// 特殊处理：检查是否为nil指针、nil切片、nil map
	if value != nil {
		valueReflect := reflect.ValueOf(value)
		kind := valueReflect.Kind()

		// 如果是nil指针、nil切片、nil map，使用特殊标记
		if (kind == reflect.Ptr || kind == reflect.Slice || kind == reflect.Map) && valueReflect.IsNil() {
			// 使用类型信息包装一下nil值
			typeName := valueReflect.Type().String()
			nilMarker := &nilValueMarker{TypeName: typeName}

			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			// 使用与Encode一致的方式：编码interface{}的指针
			var markerInterface interface{} = nilMarker
			if err := enc.Encode(&markerInterface); err != nil {
				return nil, fmt.Errorf("gob encode error: %w", err)
			}
			return buf.Bytes(), nil
		}
	}

	// 注册类型
	registerTypeIfNeeded(value)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	// 直接编码value的指针
	if err := enc.Encode(&value); err != nil {
		return nil, fmt.Errorf("gob encode error: %w", err)
	}
	return buf.Bytes(), nil
}

// Decode 使用gob反序列化
func Decode(data []byte, obj any) error {
	if obj == nil {
		return fmt.Errorf("obj cannot be nil")
	}

	// 检查obj必须是指针
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		return fmt.Errorf("obj must be a pointer")
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	// 解码到临时变量
	var value interface{}
	if err := dec.Decode(&value); err != nil {
		return fmt.Errorf("gob decode error: %w", err)
	}

	// 检查是否为nilValueMarker
	if _, ok := value.(*nilValueMarker); ok {
		// 这是一个nil值
		return assignValue(obj, nil)
	}

	// 将value赋给obj
	return assignValue(obj, value)
}

// assignValue 使用反射将值赋给目标对象
func assignValue(obj any, value interface{}) error {
	objValue := reflect.ValueOf(obj)
	objElem := objValue.Elem()

	if !objElem.CanSet() {
		return fmt.Errorf("obj cannot be set")
	}

	// 处理nil值
	if value == nil {
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

	valueReflect := reflect.ValueOf(value)
	if !valueReflect.IsValid() {
		return fmt.Errorf("invalid value")
	}

	// 类型必须匹配
	if objElem.Type() != valueReflect.Type() {
		return fmt.Errorf("type mismatch: expected %s, got %s", objElem.Type(), valueReflect.Type())
	}

	objElem.Set(valueReflect)
	return nil
}
