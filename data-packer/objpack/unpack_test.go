package objpack_test

import (
	"fmt"
	"reflect"
	"testing"

	objpack "github.com/touee/nyn/data-packer/objpack"
)

func TestUnpack(t *testing.T) {
	type Table struct {
		object interface{}
	}

	/*
		var tables = []Table{
			{"abcdefg"},
			//{new(int)},
			//{new(*int)},
		}
	*/

	type tt struct {
		A bool
		B uint
		C int
		Z struct {
			X []byte
		}
	}
	type ttt [5]byte
	var o1 = ttt{1, 55} /*tt{true, 123, 123, struct {
		X []byte
	}{[]byte("aaa")}}*/
	var packed []byte
	var err error
	packed, err = objpack.Pack(o1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", packed)
	var u1 ttt
	err = objpack.Unpack(packed, &u1)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(u1)

	/*
		for _, table := range tables {
			var packedData, err = objpack.Pack(table.object)
			if err != nil {
				t.Error("packing", err)
				continue
			}

			var unpacked = reflect.Zero(reflect.TypeOf(table.object)).Interface()
			err = objpack.Unpack(packedData, &unpacked)
			if err != nil {
				t.Error("unpacking", err)
				continue
			}

			if !reflect.DeepEqual(table.object, unpacked) {
				t.Errorf("not equal! %v != %v !", table.object, unpacked)
			}

			t.Logf("%v", unpacked)
		}
	*/
}

func TestSomething(t *testing.T) {
	type EmbeddedStruct struct {
		SliceField []bool
		//MapField   map[string]*EmbeddedStruct
		MapField map[string]int
	}
	type Data struct {
		Float64Field float64
		StringField  string
		StructField  EmbeddedStruct
	}

	//var originalData = Data{123.4, "Hello", EmbeddedStruct{[]bool{true, false}, map[string]*EmbeddedStruct{"World": &EmbeddedStruct{}, "!": &EmbeddedStruct{}}}}
	var originalData = Data{123.4, "Hello", EmbeddedStruct{[]bool{true, false}, map[string]int{"World": 1, "!": 2}}}

	/*
		var signature, errSignature = objpack.MakeTypeSignature(originalData)
		if errSignature != nil {
			panic(errSignature)
		}
		fmt.Println(signature)
	*/

	var packedData, errPack = objpack.Pack(originalData)
	if errPack != nil {
		panic(errPack)
	}
	fmt.Printf("%v\n", packedData)

	var unpackedDataPointer = reflect.New(reflect.TypeOf(Data{})).Interface()
	var errUnpack = objpack.Unpack(packedData, unpackedDataPointer)
	if errUnpack != nil {
		panic(errUnpack)
	}
	var unpackedData = *unpackedDataPointer.(*Data)
	if !reflect.DeepEqual(originalData, unpackedData) {
		fmt.Println(originalData.StructField.MapField["!"], unpackedData.StructField.MapField["!"])
		t.Errorf("%#v ≠ %#v", originalData, unpackedData)
	}
}
