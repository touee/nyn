package objpack

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

// Unpack unpacks the encoded data and stores the result in the value pointed to by object.
func Unpack(data []byte, object interface{}) (err error) {
	return UnpackFromReader(bytes.NewReader(data), reflect.ValueOf(object))
}

// UnpackFromReader reads the encoded data from reader and unpacks and stores the result in the value pointed to by v.
func UnpackFromReader(reader *bytes.Reader, v reflect.Value) (err error) {
	var t = v.Type()
	if t.Kind() != reflect.Ptr || v.IsNil() {
		return BadUnpackTypeError{t}
	} /*else if hasCircling(t) {
		return ErrCircling{t}
	}*/

	defer func() {
		if p := recover(); p != nil {
			err = PanicInUnpackingError{p}
		}
	}()

	return unpackFromReader(reader, reflect.Indirect(v))
}

func unpackFromReader(reader *bytes.Reader, v reflect.Value) (err error) {
	var t = v.Type()
	var kind = t.Kind()

	if kind == reflect.Interface {
		v = v.Elem()
		t = v.Type()
		kind = t.Kind()
	}

	switch kind {
	case reflect.Ptr:
		var vElemPointer = reflect.New(t.Elem())
		err = unpackFromReader(reader, vElemPointer.Elem())
		if err != nil {
			return err
		}
		v.Set(vElemPointer)
	case reflect.Struct:
		for i, num := 0, v.NumField(); i < num; i++ {
			var structField = t.Field(i)
			if structField.PkgPath == "" /* && structField.Tag.Get("binpacck") != "-" */ {
				var field = v.Field(i)
				err = unpackFromReader(reader, field)
				if err != nil {
					return err
				}
			} else {
				return UnexportedFieldError{structField}
			}
		}
	case reflect.Bool, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Float32, reflect.Float64:
		err = binary.Read(reader, binary.LittleEndian, v.Addr().Interface())
		if err != nil {
			return err
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		var i int64
		i, err = binary.ReadVarint(reader)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		var u uint64
		u, err = binary.ReadUvarint(reader)
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.String:
		var n int64
		n, err = binary.ReadVarint(reader)
		if err != nil {
			return err
		}
		var buf = make([]byte, n)
		_, err = reader.Read(buf)
		if err != nil {
			return err
		}
		v.SetString(string(buf))
	case reflect.Slice:
		var _count int64
		_count, err = binary.ReadVarint(reader)
		var count = int(_count)
		if err != nil {
			return err
		}
		var vSlice = reflect.MakeSlice(t, int(count), int(count))
		for i := 0; i < count; i++ {
			err = unpackFromReader(reader, vSlice.Index(i))
			if err != nil {
				return err
			}
		}
		v.Set(vSlice)
	case reflect.Array:
		var count = v.Len()
		for i := 0; i < count; i++ {
			err = unpackFromReader(reader, v.Index(i))
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		var n int64
		n, err = binary.ReadVarint(reader)
		if err != nil {
			return err
		}
		var m = reflect.MakeMap(reflect.MapOf(t.Key(), t.Elem()))
		for i := 0; i < int(n); i++ {
			var vKey, vValue = reflect.New(t.Key()).Elem(), reflect.New(t.Elem()).Elem()
			unpackFromReader(reader, vKey)
			unpackFromReader(reader, vValue)
			m.SetMapIndex(vKey, vValue)
		}
		v.Set(m)
	case reflect.Uintptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		return UnresolvableTypeError{t}
	case reflect.Complex64, reflect.Complex128:
		// TODO: Support them
		fallthrough
	default:
		return UnsupportedTypeError{t}
	}
	return nil
}
