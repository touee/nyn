package objpack_test

import (
	"testing"

	objpack "github.com/touee/nyn/data-packer/objpack"
)

func TestPack(t *testing.T) {
	type Table struct {
		object interface{}
	}

	var tables = []Table{
		{"abcdefg"},
		{new(int)},
		{new(*int)},
	}

	type S2 struct {
		f []bool
	}
	//tables = append(tables, Table{S2{}})

	type S4 struct {
		F []bool
	}
	type S3 struct {
		F *S4
	}

	tables = append(tables, Table{S3{F: &S4{[]bool{true, false}}}})
	tables = append(tables, Table{S3{F: &S4{nil}}})

	for _, table := range tables {
		var result, err = objpack.Pack(table.object)
		if err != nil {
			t.Error("error:", err)
			continue
		} else {
			t.Logf("no errors: %#v = %v", table.object, result)
		}
	}

}
