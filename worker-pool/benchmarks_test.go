package workerpool_test

import (
	"log"
	"sync"
	"testing"

	workerpool "github.com/touee/nyn/worker-pool"
	workerpool_old "github.com/touee/nyn/worker-pool/old"
)

func BenchmarkOldPool(b *testing.B) {
	var table = tableTestWorkerPool{100, b.N, -1}
	testOldWorkerPool(b, table, 1)
}

func BenchmarkNewPool(b *testing.B) {
	var table = tableTestWorkerPool{100, b.N, -1}
	testWorkerPool(b, table, 1)
}

func testOldWorkerPool(t interface{ Fail() }, table tableTestWorkerPool, count int) {
	for j := 0; j < count; j++ {
		log.Printf("SUB %d", j)
		var p = workerpool_old.NewWorkerPool(table.poolSize)
		var s = make([]bool, table.taskCount)

		var count int
		var countLock sync.Mutex
		go func() {
			for i := 0; i < len(s); i++ {
				if table.resizedPoolSize >= 0 && i == len(s)/2 {
					//log.Printf("WILL CALL SetWorkerCapacity(%d)", table.resizedPoolSize)
					var err = p.SetCapacity(table.resizedPoolSize)
					//log.Printf("DID CALL SetWorkerCapacity(%d)", table.resizedPoolSize)
					if table.resizedPoolSize == 0 && err != workerpool.ErrInvaildWorkerCapacity {
						t.Fail()
					}
				}

				p.AddWorkSync(func(i int) func() {
					return func() {
						countLock.Lock()
						//t.Log(i)
						//log.Println(i)
						count++
						countLock.Unlock()
						//time.Sleep(time.Millisecond * 50)

						s[i] = true
					}
				}(i))

			}
			p.NoMoreWorks()
		}()

		p.WaitQuit()

		for i := 0; i < len(s); i++ {
			if !s[i] {
				t.Fail()
			}
		}
	}
}
