package workerpool_old

import (
	"sync"
)

// WorkerPool 是 worker pool
type WorkerPool struct {
	workerCapacity, remainWorkers int
	workersToBeQuitted            int
	lock                          sync.Mutex

	externalMessageChan chan interface{}

	cOut chan interface{}
	cIn  chan interface{}

	quitNotifyChan chan struct{}
}

// NewWorkerPool 返回一个新的 WorkerPool
func NewWorkerPool(workerCount int) (pool *WorkerPool) {

	if workerCount < 0 {
		panic("worker capacity must > 0")
	}

	pool = new(WorkerPool)
	pool.workerCapacity = workerCount
	pool.remainWorkers = pool.workerCapacity
	pool.externalMessageChan = make(chan interface{})
	pool.cOut, pool.cIn = make(chan interface{}), make(chan interface{})

	pool.quitNotifyChan = make(chan struct{})

	// 假装一开始所有 worker 就已经有任务了, worker 函数进入循环前也有对应完成任务的通知代码
	for i := 0; i < pool.workerCapacity; i++ {
		go pool.worker(pool.cOut, pool.cIn)
	}

	go func() {
		for x := range pool.cIn {
			//log.Printf("pool\t\tRECEVIED\tmessage=%s workerCapacity=%d remainWorkers=%d workersToBeQuitted=%d", x, pool.workerCapacity, pool.remainWorkers, pool.workersToBeQuitted) // debugging
			switch x.(type) {
			case workerToPoolMessage:
				switch x {
				case workerQuitted:
					/*
						if pool.WorkChan == nil { // debugging
							print()
						}
					*/
					pool.lock.Lock()
					//log.Println("poolCommandWorkerQuitted", pool.remainWorkers) // debugging
					pool.remainWorkers--
					pool.lock.Unlock()
					if pool.remainWorkers == 0 {
						//log.Printf("pool\t\tQUITTING") // debugging
						close(pool.cIn)
						close(pool.cOut)
						pool.quitNotifyChan <- struct{}{}
						close(pool.quitNotifyChan)
						//log.Printf("pool\t\tQUITTED") // debugging
						return
					}
				case workerDoneTask:

					pool.lock.Lock()
					//log.Println("poolCommandWorkDone", pool.workersToBeQuitted) // debugging
					if pool.workersToBeQuitted > 0 {
						pool.workersToBeQuitted--
						var c = workerShouldQuit
						pool.cOut <- c
						//log.Printf("pool\t\tSENT\t\tmessage=%s", c) // debugging
					} else if pool.externalMessageChan != nil {

						var msg = <-pool.externalMessageChan
						switch msg.(type) {
						case nil:
							close(pool.externalMessageChan)
							pool.externalMessageChan = nil
							pool.setCapacity(0, true)
							pool.cOut <- workerShouldDoNothing
						case addWorkMessage:
							pool.cOut <- (func())(msg.(addWorkMessage))
							//log.Printf("pool\t\tSENT\t\tmessage=%s", msg) // debugging
						case setCapacityMessage:
							pool.setCapacity(int(msg.(setCapacityMessage)), true)
							pool.cOut <- workerShouldDoNothing
						}

					} else {
						panic("!")
					}
					pool.lock.Unlock()
				}
			default:
				panic("?")
			}
		}
	}()

	return pool
}

//var debugIDFactory int     // debugging
//var debugIDLock sync.Mutex // debugging

func (pool *WorkerPool) worker(cIn <-chan interface{}, cOut chan<- interface{}) {
	cOut <- workerDoneTask

	/*
		debugIDLock.Lock()           // debugging
		var debugID = debugIDFactory // debugging
		debugIDFactory++             // debugging
		debugIDLock.Unlock()         // debugging
	*/

	for x := range cIn {
		//log.Printf("worker[%d]\tRECEIVED\tmessage=%s", debugID, x) // debugging
		switch x.(type) {
		case poolToWorkerMessage:
			switch x {
			case workerShouldQuit:
				// 因收到单个退出命令而退出
				var c = workerQuitted
				//log.Printf("worker[%d]\tWILL SEND\tmessage=%s", debugID, c) // debugging
				cOut <- c
				//log.Printf("worker[%d]\tSENT\t\tmessage=%s", debugID, c) // debugging
				//log.Printf("worker[%d]\tQUITTED", debugID)               // debugging
				return
			case workerShouldDoNothing:
				cOut <- workerDoneTask
			default:
				panic("?")
			}
		case func():
			x.(func())()
			var c = workerDoneTask
			//log.Printf("worker[%d]\tWILL SEND\tmessage=%s", debugID, c) // debugging
			cOut <- c
			//log.Printf("worker[%d]\tSENT\t\tmessage=%s", debugID, c) // debugging
		default:
			panic("?")
		}
	}
	// 因关闭管道而退出 (因收到全体退出命令而退出)
	//cOut <- poolCommandWorkerQuitted
	//panic("?")
}

// SetCapacity 设置 worker pool 的 worker 容量
func (pool *WorkerPool) SetCapacity(i int) (err error) {
	if i <= 0 {
		return ErrInvaildWorkerCapacity
	}
	pool.externalMessageChan <- setCapacityMessage(i)
	return nil
}

func (pool *WorkerPool) setCapacity(i int, alreadyLocked bool) {

	if !alreadyLocked {
		pool.lock.Lock()
	}
	if pool.workerCapacity == 0 {
		// 此时, pool 已决定要退出
		if !alreadyLocked {
			pool.lock.Unlock()
		}
		return
	}

	if i > pool.workerCapacity {
		// 现有 worker 少于设置值时, 创建新的 worker
		for ; pool.workerCapacity < i; pool.workerCapacity++ {
			pool.remainWorkers++
			go pool.worker(pool.cOut, pool.cIn)
		}
	} else {
		pool.workersToBeQuitted += (pool.workerCapacity - i)
		pool.workerCapacity = i
	}

	if !alreadyLocked {
		pool.lock.Unlock()
	}
}

// AddWorkSync 添加一个工作
func (pool *WorkerPool) AddWorkSync(work func()) {
	if work != nil {
		pool.externalMessageChan <- addWorkMessage(work)
	}
}

// NoMoreWorks 告知 pool 没有更多任务
func (pool *WorkerPool) NoMoreWorks() {
	pool.externalMessageChan <- nil
}

// WaitQuit 等待 pool 的退出
func (pool *WorkerPool) WaitQuit() {
	<-pool.quitNotifyChan
}
