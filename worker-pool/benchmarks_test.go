package workerpool_test

import (
	"testing"
)

func BenchmarkNewPool(b *testing.B) {
	var table = tableTestWorkerPool{100, b.N, -1}
	testWorkerPool(b, table, 1)
}
