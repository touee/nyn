package objpack

import (
	"unsafe"
)

var isLittleEndian = false

func init() {
	var x uint16 = 0xFF00
	isLittleEndian = *((*uint8)(unsafe.Pointer(&x))) == 0
}

// unused
func reverse(data []byte) {
	if isLittleEndian {
		return
	}
	var l = len(data)
	for i := 0; i < l/2; i++ {
		data[i], data[l-i] = data[l-i], data[i]
	}
}
