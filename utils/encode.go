package utils

import (
	"errors"
	"reflect"
	"unsafe"
)

var (
	ErrUnsupportedType error = errors.New("byteconv: unsupported type")
)

func ToBytes(val interface{}) ([]byte, error) {
	return encoderSwitch(val)
}

func encoderSwitch(val interface{}) ([]byte, error) {
	kind := reflect.TypeOf(val).Kind()
	switch kind {
	case reflect.Bool:
		return boolEncoder(val.(bool)), nil
	case reflect.String:
		return stringEncoder(val.(string)), nil
	case reflect.Float64:
		return floatEncoder(val.(float64)), nil
	case reflect.Float32:
		return floatEncoder(float64(val.(float32))), nil
	case reflect.Int:
		return intEncoder(val.(int)), nil
	case reflect.Int8:
		return intEncoder(int(val.(int8))), nil
	case reflect.Int16:
		return intEncoder(int(val.(int16))), nil
	case reflect.Int32:
		return intEncoder(int(val.(int32))), nil
	case reflect.Int64:
		return intEncoder(int(val.(int64))), nil
	case reflect.Uint:
		return intEncoder(int(val.(uint))), nil
	case reflect.Uint8:
		return intEncoder(int(val.(uint8))), nil
	case reflect.Uint16:
		return intEncoder(int(val.(uint16))), nil
	case reflect.Uint32:
		return intEncoder(int(val.(uint32))), nil
	case reflect.Uint64:
		return intEncoder(int(val.(uint64))), nil
	}
	return nil, ErrUnsupportedType
}

func boolEncoder(val bool) []byte {
	if val {
		return []byte{1}
	}
	return []byte{0}
}

func stringEncoder(val string) []byte {
	return *(*[]byte)(unsafe.Pointer(&val))
}

func intEncoder(val int) []byte {
	size := int(unsafe.Sizeof(val))
	point := uintptr(unsafe.Pointer(&val))
	return coreEncoder(size, point)
}

func floatEncoder(val float64) []byte {
	size := int(unsafe.Sizeof(val))
	point := uintptr(unsafe.Pointer(&val))
	return coreEncoder(size, point)
}

func coreEncoder(size int, point uintptr) []byte {
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(point + uintptr(i)))
		arr[i] = byt
	}
	return arr
}
