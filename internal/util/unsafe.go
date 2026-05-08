package util

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Ptr returns the pointer to these things:
//   - First element of a slice or array, nil if it has zero length;
//   - The pointer itself if the parameter is a pointer/uintptr, pointing to an scalar value.
// It panics otherwise.
func Ptr(d interface{}) (addr unsafe.Pointer) {
	if d == nil {
		return unsafe.Pointer(nil)
	}

	v := reflect.ValueOf(d)
	switch v.Type().Kind() {
	case reflect.Ptr:
		e := v.Elem()
		switch e.Kind() {
		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			addr = unsafe.Pointer(e.UnsafeAddr())
		default:
			panic(fmt.Errorf("unsupported pointer to type %s; must be a slice or pointer to a singular scalar value or the first element of an array or slice", e.Kind()))
		}

	case reflect.Slice, reflect.Array:
		if v.Len() != 0 {
			addr = unsafe.Pointer(v.Index(0).UnsafeAddr())
		}
	case reflect.Uintptr:
		addr = unsafe.Pointer(d.(uintptr))
	default:
		panic(fmt.Errorf("unsupported type %s; must be a slice or pointer to a singular scalar value or the first element of an array or slice", v.Type()))
	}
	return
}
