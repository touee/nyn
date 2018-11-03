package sscheduler

import (
	"github.com/touee/nyn/logger"
	"github.com/touee/nyn/scheduler"
	taskqueue "github.com/touee/nyn/task-queue"
	typemanager "github.com/touee/nyn/type-manager"
)

// Run 启动调度器
func (s *SimpleScheduler) Run() {
	go s.runSync()
}

func (s *SimpleScheduler) runSync() {
	var ok bool
	for {
		//s.log(logger.LDebug, "New turn.", nil)
		if ok = s.workerPool.AddWorkIfCouldSync(func() (work func()) {
			var err error
			var (
				taskID   int
				typeName string
				taskData []byte
			)
			taskID, typeName, taskData, err = s.taskQueue.DequeueForProcess()
			if err == taskqueue.ErrNoTasksInPending { //< 没有等待中的任务
				if s.mode == scheduler.ModeCollector {
					s.log(logger.LInfo, "no tasks in pending in this turn!", nil)
					return nil
				}
				panic("not implemented")
			} else if err != nil {
				panic(err)
				//s.log(logger.LError, "Dequeuing encountered error.", logger.Fields{{"error", err}})
			}

			return func(taskID int, typeName string, taskData []byte) func() {
				return func() {
					// 这里应该要保证不会发生错误, 因为若发生了错误的话, 没有办法去处理该错误
					// 但是发生了错误基本也会是 taskHandler 返回了不符规则的处理结果所致…
					var err error
					var task interface{}
					task, err = s.typeManager.Unpack(typeName, taskData)
					if err != nil {
						s.log(logger.LError, "unpacking dequeued task encountered error.", logger.Fields{{"error", err.Error()}})
					}
					var result = s.taskHandler(task)
					if err = s.taskQueue.ReportProcessResult(taskID, result); err != nil {
						s.log(logger.LError, "reporting task's process result to task queue encountered error.", logger.Fields{{"error", err}})
					}
				}
			}(taskID, typeName, taskData)

		}); !ok {
			s.log(logger.LInfo, "worker pool has been closed, quit run loop.", nil)
			break
		}
	}
}

// WaitQuit 通知调度器关闭并且等待其完成关闭
func (s *SimpleScheduler) WaitQuit() (errs []error) {

	var err error

	//var commandChan = s.commandChan
	//s.commandChan = nil
	//close(commandChan)
	s.workerPool.WaitQuit()
	if err = s.typeManager.Close(); err != nil {
		s.log(logger.LError, "close type manager encountered error.", logger.Fields{{"error", err}})
		errs = append(errs, err)
	}
	if err = s.taskQueue.Close(); err != nil {
		s.log(logger.LError, "close task queue encountered error.", logger.Fields{{"error", err}})
		errs = append(errs, err)
	}
	return errs
}

// AddTask 添加一个任务
func (s *SimpleScheduler) AddTask(task interface{}, opts scheduler.AddTaskOptions) (err error) {
	s.log(logger.LTrace, "task added.", logger.Fields{{"task-type", typemanager.GetTypeName(task)}, {"task", task}})
	var taskType = typemanager.GetTypeName(task)
	var taskData []byte
	if taskData, err = s.typeManager.Pack(task); err != nil {
		return err
	}

	if _, err = s.taskQueue.EnqueueWithOptions(taskType, taskData, opts); err != nil {
		return err
	}
	return nil
}
