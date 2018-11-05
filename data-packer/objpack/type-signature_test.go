package objpack_test

import (
	"testing"

	objpack "github.com/touee/nyn/data-packer/objpack"
)

func TestSignature(t *testing.T) {
	type Table struct {
		object interface{}
		result string
	}

	var tables = []Table{
		{int(1), "int"},
		{new(int), "int"},
		{[1]byte{1}, "[1]uint8"},
		{make([]struct{ a **string }, 0), "[]struct{a:string}"},
		{struct{ a string }{}, "struct{a:string}"},
		{struct{ a struct{ b []bool } }{}, "struct{a:struct{b:[]bool}}"},
	}

	for _, table := range tables {
		var result, err = objpack.MakeTypeSignature(table.object)
		if err != nil {
			t.Errorf("input: %v, expect: %s, error: %s", table.object, table.result, err)
		} else if result != table.result {
			t.Errorf("input: %v, expect: %s, get: %s", table.object, table.result, result)
		}
	}
}
