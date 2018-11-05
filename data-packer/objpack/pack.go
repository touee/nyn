package objpack

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"sort"
)

// Pack returns the packed data of object.
func Pack(object interface{}) (result []byte, err error) {
	var buf bytes.Buffer
	err = PackToWriter(&buf, reflect.ValueOf(object))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil

}

// PackToWriter packs and writes the packed data of object to writer
func PackToWriter(writer io.Writer, v reflect.Value) (err error) {

	/*
		var t = v.Type()
		if hasCircling(t) {
			return ErrCircling{t}
		}
	*/

	return packToWriter(writer, v)
}

func packToWriter(writer io.Writer, v reflect.Value) (err error) {
	var t reflect.Type
	var kind reflect.Kind

	for {
		t = v.Type()
		kind = t.Kind()
		if kind == reflect.Ptr {
			if v.IsNil() {
				v = reflect.New(t.Elem())
			} else {
				v = reflect.Indirect(v)
			}
		} else if kind == reflect.Slice && v.IsNil() {
			v = reflect.MakeSlice(t, 0, 0)
		} else {
			break
		}
	}

	var vBuf = make([]byte, binary.MaxVarintLen64)
	switch kind {
	case reflect.Struct:
		for i, num := 0, v.NumField(); i < num; i++ {
			var structField = t.Field(i)
			if structField.PkgPath == "" /* && structField.Tag.Get("binpacck") != "-" */ {
				var field = v.Field(i)
				err = packToWriter(writer, field)
				if err != nil {
					return err
				}
			} else {
				return UnexportedFieldError{structField}
			}
		}
	case reflect.Bool, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Float32, reflect.Float64:
		binary.Write(writer, binary.LittleEndian, v.Interface())
	case reflect.Int, reflect.Int32, reflect.Int64:
		var n = binary.PutVarint(vBuf, v.Int())
		writer.Write(vBuf[:n])
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		var n = binary.PutUvarint(vBuf, v.Uint())
		writer.Write(vBuf[:n])
	case reflect.String:
		var str = v.String()
		var n = binary.PutVarint(vBuf, int64(len(str)))
		writer.Write(vBuf[:n])
		io.WriteString(writer, str)
	case reflect.Slice, reflect.Array:
		var count = v.Len()
		if kind == reflect.Slice { //< 只有 slice 放长度
			//var elemSize = int(v.Type().Elem().Size())
			//var byteSize = count * elemSize
			//var n = binary.PutVarint(vBuf, int64(byteSize))
			var n = binary.PutVarint(vBuf, int64(count))
			writer.Write(vBuf[:n])
		}

		for i := 0; i < count; i++ {
			err = packToWriter(writer, v.Index(i))
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		var n = binary.PutVarint(vBuf, int64(v.Len()))
		writer.Write(vBuf[:n])
		var kvBufs = make([][2][]byte, v.Len())
		for i, key := range v.MapKeys() {
			var keyBuf, valueBuf bytes.Buffer
			err = packToWriter(&keyBuf, key)
			if err != nil {
				return err
			}
			err = packToWriter(&valueBuf, v.MapIndex(key))
			if err != nil {
				return err
			}
			kvBufs[i][0], kvBufs[i][1] = keyBuf.Bytes(), valueBuf.Bytes()
		}
		sort.Sort(kvBufsSortor(kvBufs))
		for _, x := range kvBufs {
			writer.Write(x[0])
			writer.Write(x[1])
		}
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

type kvBufsSortor [][2][]byte

func (s kvBufsSortor) Len() int {
	return len(s)
}

func (s kvBufsSortor) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s kvBufsSortor) Less(i, j int) bool {
	return bytes.Compare(s[i][0], s[j][0]) < 0
}
