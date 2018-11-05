# objpack

A tiny and stupid object packing/unpacking package written in Golang for my personal project

## Limits and Features

### Limits

* All fields in a struct must be exported.
* All your data's types must be explicit.
* To unpack data properly, the types' signature of original object and target object must match.
  * This means once you packed your object, you'd better not change the structure for your data, or you will need to convert them manually.
* complex64/complex128 has not been supported.
* circular references will cause stack overflow.

### Supported Types

* basic types
  * `bool`
  * `uint` `uint8` `uint16` `uint32` `uint64` `byte`
  * `int` `int8` `int16` `int32` `int64`
  * `float32` `float64`
  * `string`

* arrays, slices and maps of supported type
* structs that all fields are exported, and all fields' type has been supported

## Example

```go
type EmbeddedStruct struct {
    SliceField []*bool
    MapField   map[string]uint
}
type Data struct {
    Float64Field float64
    StringField  string
    StructField  EmbeddedStruct
}

var originalData = Data{123.4, "Hello", EmbeddedStruct{[]*bool{new(bool), new(bool)}, map[string]uint{"World": 0, "!": 1}}}

var signature, errSignature = objpack.MakeTypeSignature(originalData)
if errSignature != nil {
    panic(errSignature)
}
fmt.Println(signature)

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
fmt.Println(reflect.DeepEqual(originalData, unpackedData))
```

## Todo

* [ ] Handling nil
* [ ] Comments/Documents
* [ ] Packer/Unpacker
* [ ] Benchmark
* [ ] Improve Performance 