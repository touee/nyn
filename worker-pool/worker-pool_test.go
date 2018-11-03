package workerpool_test

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/touee/nyn/logger"

	"github.com/touee/nyn/worker-pool"
)

type tableTestWorkerPool struct {
	poolSize        int
	taskCount       int
	resizedPoolSize int
}

func TestWorkerPool(t *testing.T) {

	for i, table := range []tableTestWorkerPool{
		{6, 3, -1},
		{100, 1000, -1},
		{6, 1000, 0},
		{6, 1000, 12},
		{12, 1000, 6}, //< deadlocked
	} {
		_ = i
		log.Println(fmt.Sprintf("BEGIN: %d %#v", i, table))
		testWorkerPool(t, table, 100)
		log.Println(fmt.Sprintf("OK: %d %#v", i, table))
	}

}

func testWorkerPool(t interface{ Fail() }, table tableTestWorkerPool, count int) {
	for j := 0; j < count; j++ {
		log.Printf("SUB %d", j)
		var p = workerpool.NewWorkerPool(table.poolSize)
		p.SetLogger(logger.NewWriterLogger(os.Stdout, logger.LDebug)) 
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

				if p.AddWorkIfCouldSync(func(i int) func() func() {
					return func() func() {
						return func() {
							countLock.Lock()
							//t.Log(i)
							//log.Println(i)
							count++
							countLock.Unlock()

							//time.Sleep(time.Millisecond * 50)
							s[i] = true
						}
					}
				}(i)) == false {
					panic(false)
				}
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
