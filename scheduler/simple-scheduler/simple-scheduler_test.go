package sscheduler_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/touee/nyn/logger"
	"github.com/touee/nyn/scheduler"
	"github.com/touee/nyn/scheduler/simple-scheduler"
	"github.com/touee/nyn/task-queue"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func randomDir() string {
	//return path.Join(os.TempDir(), fmt.Sprintf("_temp_sscheduler_%d", rand.Int()))
	return fmt.Sprintf("_temp_sscheduler_%d", rand.Int())
}

func TestCollectorMode(t *testing.T) {
	var err error

	var dir = randomDir()
	if err = os.MkdirAll(dir, 0644); err != nil {
		t.Fatal(err)
	}
	//defer os.RemoveAll(dir)

	var s scheduler.Scheduler
	if s, err = sscheduler.NewSimpleScheduler(sscheduler.SimpleSchedulerOptions{
		Dir:                      dir,
		DefaultWorkerPoolWorkers: 10,
	}); err != nil {
		t.Fatal(err)
	}

	type TestTask struct {
		Index int
		Step  int
	}

	const (
		b             = 1 << 12
		failCondition = 7
	)

	var (
		expectedFailCount = b / failCondition
		area              = make([]bool, b)
	)

	var opts = scheduler.AddTaskOptions{Priority: 1, MaxAttempts: 1}

	s.SetMode(scheduler.ModeCollector)
	s.SetLogger(logger.NewWriterLogger(os.Stdout, logger.LTrace))
	s.SetTaskHandler(func(payload interface{}) taskqueue.ProcessResult {
		var task = payload.(TestTask)
		//log.Println(task.Index)
		if task.Index%2 != 1 {
			var (
				task1 = TestTask{task.Index - task.Step, task.Step / 2}
				task2 = TestTask{task.Index + task.Step, task.Step / 2}
			)
			//log.Printf("new tasks: %d and %d", task1.Index, task2.Index)
			s.AddTask(task1, opts)
			s.AddTask(task2, opts)
		}
		if task.Index%failCondition == 0 {
			return taskqueue.ProcessResultFailed
		}
		if area[task.Index] {
			panic("?")
		}
		area[task.Index] = true
		return taskqueue.ProcessResultSuccessful
	})

	s.RegisterTaskType(TestTask{})
	s.AddTask(TestTask{b / 2, b / 4}, opts)

	s.Run()
	s.WaitQuit()

	var count = expectedFailCount
	var areaCount int
	for _, r := range area {
		if r {
			areaCount++
		}
	}
	count += areaCount + 1 // 算上 0
	if count != b {
		t.Fatal(count, areaCount)
	}
}
