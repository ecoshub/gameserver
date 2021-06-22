package utils

import (
	"errors"
	"unsafe"
)

var (
	ErrNilArray  error = errors.New("nil array")
	ErrWrongSize error = errors.New("wrong size")
)

func ToInt(arr []byte) (int, error) {
	size := len(arr)
	var val int
	for i := 0; i < size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
	}
	return val, nil
}

func ToFloat(arr []byte) (float64, error) {
	size := len(arr)
	var val float64
	for i := 0; i < size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
	}
	return val, nil
}

func ToString(arr []byte) (string, error) {
	val := *(*string)(unsafe.Pointer(&arr))
	return val, nil
}

func ToBool(arr []byte) (bool, error) {
	lena := len(arr)
	if lena != 1 {
		return false, ErrWrongSize
	}
	if arr[0] == 1 {
		return true, nil
	}
	return false, nil
}
