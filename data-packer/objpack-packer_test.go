package datapacker_test

import (
	"fmt"
	"reflect"
	"testing"

	datapacker "github.com/touee/nyn/data-packer"
)

func TestPackAndUnpack(t *testing.T) {

	type TT1 struct {
		X int
		Y string
	}

	type TT2A struct {
		A1 int
	}

	type TT2 struct {
		As []TT2A
	}

	type Table struct {
		task              interface{}
		expectedSignature string
	}

	var p = datapacker.DefaultPacker

	if pf, sf := p.Formats(); pf != "objpack" || sf != "objpack-signature" {
		t.Fail()
	}

	for _, table := range []Table{
		{TT1{X: 1, Y: "a"}, "struct{X:int;Y:string}"},
		{TT2{As: []TT2A{
			{1}, {2}, {3},
		}}, "struct{As:[]struct{A1:int}}"},
		{TT2{As: nil}, "struct{As:[]struct{A1:int}}"},
	} {
		var err error

		var signature string
		signature, err = p.Signature(table.task)
		if err != nil {
			t.Fatal(err)
		}
		if signature != table.expectedSignature {
			t.Fatal(signature, "≠", table.expectedSignature)
		}

		var packed []byte
		packed, err = p.Pack(table.task)
		if err != nil {
			t.Fatal(err)
		}
		var unpackedP = reflect.New(reflect.TypeOf(table.task))
		err = p.Unpack(packed, unpackedP.Interface())
		if err != nil {
			t.Fatal(err)
		}
		var unpacked = reflect.Indirect(unpackedP).Interface()
		if fmt.Sprintf("%v", table.task) != fmt.Sprintf("%v", unpacked) {
			//if !reflect.DeepEqual(task, unpacked) {
			t.Fatal(table.task, "≠", unpacked)
		}
	}
}
