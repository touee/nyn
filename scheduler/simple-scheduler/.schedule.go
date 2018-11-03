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

	if s.commandChan != nil {
		panic("Duplicated runnings!")
	}

	s.commandChan = make(chan interface{})

	var err error

	for {
		var command interface{}
		if s.waitingCommand { //< 根据队列是否为空来决定是否要阻塞等待命令
			command = <-s.commandChan
		} else {
			select {
			case command = <-s.commandChan:
			default:
			}
		}

		switch command.(type) {
		case commandStop: //< 收到命令退出
			s.log(logger.LInfo, "Quit due to external command.", nil)
			s.quit()
			return
		case commandTaskAdded: //< 收到命令添加任务
			s.waitingCommand = false //< 有了任务后, 让循环不再阻塞等待命令
		case nil: //< 没有收到任何任务, 处理任务
			var (
				taskID   int
				typeName string
				taskData []byte
			)
			taskID, typeName, taskData, err = s.taskQueue.DequeueForProcess()
			if err == taskqueue.ErrNoTasksInPending { //< 没有等待中的任务
				if hadWorkersInBusy := s.workerPool.TryWaitAtLeastOneWorkDone(); hadWorkersInBusy {
					continue
				} else if taskID, typeName, taskData, err = s.taskQueue.DequeueForProcess(); err == taskqueue.ErrNoTasksInPending {
					if s.mode == scheduler.ModeCollector {
						s.log(logger.LInfo, "Quit due to no pending tasks in queue.", nil)
						s.workerPool.NoMoreWorks()
						s.quit()
						return
					}
					s.waitingCommand = true
					continue
				}
			}

			if err != nil && err != taskqueue.ErrNoTasksInPending {
				s.log(logger.LError, "Dequeuing encountered error.", logger.Fields{{"error", err}})
				continue
			}

			// 将对任务的进一步处理放在 worker pool 中去做
			go s.workerPool.AddWorkSync(func(taskID int, typeName string, taskData []byte) func() {
				return func() {
					// 这里应该要保证不会发生错误, 因为若发生了错误的话, 没有办法去处理该错误
					// 但是发生了错误基本也会是 taskHandler 返回了不符规则的处理结果所致…
					var err error
					var task interface{}
					task, err = s.typeManager.Unpack(typeName, taskData)
					if err != nil {
						s.log(logger.LError, "Unpacking dequeued task encountered error.", logger.Fields{{"error", err}})
					}
					var result = s.taskHandler(task)
					if err = s.taskQueue.ReportProcessResult(taskID, result); err != nil {
						s.log(logger.LError, "Reporting task's process result to task queue encountered error.", logger.Fields{{"error", err}})
					}
				}
			}(taskID, typeName, taskData))
		default:
			panic("?")
		}

	}
}

func (s *SimpleScheduler) quit() {
	var err error
	var errs []error

	var commandChan = s.commandChan
	s.commandChan = nil
	close(commandChan)
	s.workerPool.WaitQuit()
	if err = s.typeManager.Close(); err != nil {
		s.log(logger.LError, "Close type manager encountered error.", logger.Fields{{"error", err}})
		errs = append(errs, err)
	}
	if err = s.taskQueue.Close(); err != nil {
		s.log(logger.LError, "Close task queue encountered error.", logger.Fields{{"error", err}})
		errs = append(errs, err)
	}
	s.quitWaitChan <- errs
	close(s.quitWaitChan)
}

// AddTask 添加一个任务
func (s *SimpleScheduler) AddTask(task interface{}, opts scheduler.AddTaskOptions) (err error) {
	var taskType = typemanager.GetTypeName(task)
	var taskData []byte
	if taskData, err = s.typeManager.Pack(task); err != nil {
		return err
	}

	if _, err = s.taskQueue.EnqueueWithOptions(taskType, taskData, opts); err != nil {
		return err
	}
	if s.commandChan != nil && s.waitingCommand {
		go func() {
			s.commandChan <- commandTaskAdded{}
		}()
	}
	return nil
}
