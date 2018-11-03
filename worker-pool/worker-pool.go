package workerpool

import (
	"fmt"
	"sync"

	"github.com/touee/nyn/logger"
)

// WorkerPool 是 worker pool
type WorkerPool struct {
	workerCapacity, remainWorkers int
	workersToBeQuitted            int
	lock                          sync.Mutex

	externalMessageChan chan interface{}
	externalWaitChan    chan bool

	externalLock sync.Mutex

	cOut chan interface{}
	cIn  chan interface{}

	quitNotifyChan chan struct{}

	logger logger.Logger
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
	pool.externalWaitChan = make(chan bool)
	pool.cOut, pool.cIn = make(chan interface{}), make(chan interface{})

	pool.quitNotifyChan = make(chan struct{})

	// 假装一开始所有 worker 就已经有任务了, worker 函数进入循环前也有对应完成任务的通知代码
	for i := 0; i < pool.workerCapacity; i++ {
		go pool.worker(pool.cOut, pool.cIn)
	}

	go func() {

		var hangingWorkers int

		for x := range pool.cIn {
			pool.log(logger.LTrace, "pool received internal message.", logger.Fields{
				{"message", fmt.Sprintf("%s", x)},
				{"hangingWorkers", hangingWorkers},
				//{"workerCapacity", pool.workerCapacity},
				//{"remainWorkers", pool.remainWorkers},
				//{"workersToBeQuitted", pool.workersToBeQuitted},
			})
			switch x.(type) {
			case workerToPoolMessage:
				switch x {
				case workerQuitted:
					pool.lock.Lock()
					//log.Println("poolCommandWorkerQuitted", pool.remainWorkers) // debugging
					pool.remainWorkers--
					pool.lock.Unlock()
					if pool.remainWorkers == 0 {
						pool.log(logger.LTrace, "no workers remain. pool is quitting.", nil)
						//log.Printf("pool\t\tQUITTING") // debugging
						close(pool.cIn)
						close(pool.cOut)
						//pool.quitNotifyChan <- struct{}{}
						//close(pool.quitNotifyChan)
						//log.Printf("pool\t\tQUITTED") // debugging

						pool.quitNotifyChan <- struct{}{}
						//pool.externalWaitChan <- false
						close(pool.externalMessageChan)
						pool.externalMessageChan = nil
						close(pool.externalWaitChan)
						pool.externalWaitChan = nil

						pool.log(logger.LTrace, "pool quitted.", nil)
						return
					}
				case workerDoneTask:
					pool.lock.Lock()
					//log.Println("poolCommandWorkDone", pool.workersToBeQuitted) // debugging
					if pool.workersToBeQuitted > 0 {
						pool.workersToBeQuitted--
						var c = workerShouldQuit
						pool.cOut <- c
						//.Printf("pool\t\tSENT\t\tmessage=%s", c) // debugging
					} else if pool.externalMessageChan != nil {

						if pool.externalWaitChan != nil && pool.workerCapacity != 0 {
							pool.externalWaitChan <- true
						}

						var prepareToQuit = func() {
							//pool.externalLock.Lock()

							pool.setCapacity(0, true)
							//pool.externalLock.Unlock()

							// 挂起的 worker 直接发送退出命令
							for ; hangingWorkers != 0; hangingWorkers-- {
								pool.cOut <- workerShouldDoNothing
							}
						}

						var msg = <-pool.externalMessageChan
						pool.log(logger.LTrace, "pool received external message.", logger.Fields{
							{"message", fmt.Sprintf("%#v", msg)},
							{"hangingWorkers", hangingWorkers},
						})
						switch msg.(type) {
						case hangMessage:
							hangingWorkers++
							if hangingWorkers != pool.workerCapacity {
								break
							}
							pool.log(logger.LInfo, "all works are hanging, which means no more works to be done. pool will quit.", nil)
							// 没有工作可做而退出
							prepareToQuit()
						case nil: //< 外部主动要求退出
							pool.log(logger.LInfo, "pool received external quit message. pool will quit.", nil)
							hangingWorkers++
							prepareToQuit()
							pool.externalLock.Unlock()

						case addWorkMessage:
							pool.cOut <- (func())(msg.(addWorkMessage))

							// 重新激活挂起的 worker
							for ; hangingWorkers != 0; hangingWorkers-- {
								pool.cOut <- workerShouldDoNothing
							}
						case setCapacityMessage:
							pool.setCapacity(int(msg.(setCapacityMessage)), true)
							pool.cOut <- workerShouldDoNothing
						default:
							panic("?")
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

var debugIDFactory int     // debugging
var debugIDLock sync.Mutex // debugging

func (pool *WorkerPool) worker(cIn <-chan interface{}, cOut chan<- interface{}) {
	cOut <- workerDoneTask

	debugIDLock.Lock()           // debugging
	var debugID = debugIDFactory // debugging
	debugIDFactory++             // debugging
	debugIDLock.Unlock()         // debugging

	for x := range cIn {
		pool.log(logger.LTrace, fmt.Sprintf("worker[%d] received internal message.", debugID), logger.Fields{
			{"message", fmt.Sprintf("%s", x)},
		})
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
				pool.log(logger.LTrace, fmt.Sprintf("worker[%d] quitted due to command received.", debugID), nil)
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
	cOut <- workerQuitted
	pool.log(logger.LTrace, fmt.Sprintf("worker[%d] quitted due to channel has been closed.", debugID), nil)

	//panic("?")
}

// SetCapacity 设置 worker pool 的 worker 容量
func (pool *WorkerPool) SetCapacity(i int) (err error) {
	if i <= 0 {
		return ErrInvaildWorkerCapacity
	}
	pool.externalLock.Lock()
	<-pool.externalWaitChan
	pool.externalMessageChan <- setCapacityMessage(i)
	pool.externalLock.Unlock()
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

/*
// WaitForAddWork 等待到能添加工作为止 (有工作完成时), 返回是否能添加新的工作
func (pool *WorkerPool) WaitForAddWork() bool {
	if pool.externalWaitChan == nil {
		return false
	}
	return <-pool.externalWaitChan
}

// AddWorkSync 添加一个工作, 需要先调用 WaitForAddWork
func (pool *WorkerPool) AddWorkSync(work func()) {
	if work == nil {
		panic("Work should not be nil!")
	}
	pool.externalMessageChan <- addWorkMessage(work)
}
*/

// AddWorkIfCouldSync 在还能添加工作时, 添加 workProvider 所返回的工作. 返回能否添加工作
// 如果 workProvider 返回了 nil, 会视为 hang
func (pool *WorkerPool) AddWorkIfCouldSync(workProvider func() (work func())) bool {
	pool.externalLock.Lock()
	defer func() { pool.externalLock.Unlock() }()

	if _, ok := <-pool.externalWaitChan; !ok {
		return false
	}

	var work = workProvider()
	if work == nil {
		//panic("Work should not be nil!")
		pool.externalMessageChan <- hangMessage{}
	} else {
		pool.externalMessageChan <- addWorkMessage(work)
	}

	return true
}

// NoMoreWorks 告知 pool 没有更多任务
func (pool *WorkerPool) NoMoreWorks() {
	pool.externalLock.Lock()
	<-pool.externalWaitChan
	pool.externalMessageChan <- nil
	//pool.externalLock.Unlock()
}

// WaitQuit waits to quit
func (pool *WorkerPool) WaitQuit() {
	<-pool.quitNotifyChan
}
