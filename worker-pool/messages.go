package workerpool

type (
	poolToWorkerMessage int
	workerToPoolMessage int
)

const (
	workerShouldQuit poolToWorkerMessage = iota
	workerShouldDoNothing
)
const (
	workerQuitted workerToPoolMessage = iota
	workerDoneTask
)

func (c poolToWorkerMessage) String() string {
	switch c {
	case workerShouldQuit:
		return "workerShouldQuit"
	case workerShouldDoNothing:
		return "workerShouldDoNothing"
	}
	panic("?")
}
func (c workerToPoolMessage) String() string {
	switch c {
	case workerQuitted:
		return "workerQuitted"
	case workerDoneTask:
		return "workerDoneTask"
	}
	panic("?")
}

// setCapacityMessage 代表设置 pool 容量的消息
type setCapacityMessage int

// addWorkMessage 代表添加任务
type addWorkMessage func()

// hangMessage 代表让 worker 挂起
type hangMessage struct{}
