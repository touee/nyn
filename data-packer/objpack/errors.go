package objpack

import (
	"fmt"
	"reflect"
)

// UnresolvableTypeError 是遇到无法 pack/unpack 的类型时返回的错误
type UnresolvableTypeError struct {
	Type reflect.Type
}

func (e UnresolvableTypeError) Error() string {
	return fmt.Sprintf("objpack: Unresolvable type occurred: %s", e.Type.String())
}

// UnsupportedTypeError 是遇到尚未支持的类型时返回的错误
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e UnsupportedTypeError) Error() string {
	return fmt.Sprintf("objpack: Unsupported type occurred: %s", e.Type.String())
}

// UnexportedFieldError 是在结构体中遇到未导出的字段时返回的错误
type UnexportedFieldError struct {
	Field reflect.StructField
}

func (e UnexportedFieldError) Error() string {
	return fmt.Sprintf("objpack: Unexported field occurred: %s", e.Field.Name)
}

// BadUnpackTypeError 是在 unpack 时传入了非指针类型 object 参数是返回的错误
type BadUnpackTypeError struct {
	Type reflect.Type
}

func (e BadUnpackTypeError) Error() string {
	return fmt.Sprintf("objpack: The given object must be a pointer: %s", e.Type.String())
}

// PanicInUnpackingError 实在 unpack 中发生 panic 时返回的错误
type PanicInUnpackingError struct {
	Panic interface{}
}

func (e PanicInUnpackingError) Error() string {
	return fmt.Sprintf("objpack: Panic in unpacking: %s", e.Panic)
}

// ErrCircling 实在 unpack 中发生 panic 时返回的错误
type ErrCircling struct {
	Type reflect.Type
}

func (e ErrCircling) Error() string {
	return fmt.Sprintf("objpack: Has circling: %s", e.Type.String())
}
