package objpack

import (
	"bytes"
	"reflect"
	"strconv"
)

// MakeTypeSignature returns the signature of the type of given object.
func MakeTypeSignature(object interface{}) (result string, err error) {

	var t = reflect.TypeOf(object)

	/*
		if hasCircling(t) {
			return "", ErrCircling{t}
		}
	*/

	var buf bytes.Buffer
	err = makeTypeSignature(&buf, t, false)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func makeTypeSignature(buf *bytes.Buffer, t reflect.Type, withSemicolon bool) (err error) {
	var kind reflect.Kind
	for {
		kind = t.Kind()
		if kind == reflect.Slice {
			buf.Write([]byte("[]"))
		} else if kind == reflect.Array {
			buf.WriteByte('[')
			buf.WriteString(strconv.Itoa(t.Len()))
			buf.WriteByte(']')
		} else if kind != reflect.Ptr {
			break
		}
		t = t.Elem()
	}

	switch kind {
	case reflect.Uintptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		return UnresolvableTypeError{t}
	case reflect.Struct:
		buf.Write([]byte("struct{"))
		for i, num := 0, t.NumField(); i < num; i++ {
			var field = t.Field(i)
			buf.WriteString(field.Name)
			buf.WriteByte(':')
			err = makeTypeSignature(buf, field.Type, i != num-1)
			if err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case reflect.Map:
		buf.Write([]byte("map["))
		err = makeTypeSignature(buf, t.Key(), false)
		buf.WriteByte(']')
		err = makeTypeSignature(buf, t.Elem(), false)
	default:
		buf.WriteString(kind.String())
	}
	if withSemicolon {
		buf.WriteByte(';')
	}

	return nil
}
