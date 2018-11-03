package s3queue

import (
	"database/sql"

	taskqueue "github.com/touee/nyn/task-queue"
)

// DequeueForProcess 取出一个等待队列中处于等待状态的任务, 并将其标记为处理中状态
func (q *Queue) DequeueForProcess() (id int, taskType string, taskData []byte, err error) {
	q.dbLock.Lock()
	defer q.dbLock.Unlock()
	return q.dequeueForProcess()
}

func (q *Queue) dequeueForProcess() (id int, taskType string, taskData []byte, err error) {

	var tx *sql.Tx
	if tx, err = q.db.Begin(); err != nil {
		return 0, "", nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var taskTypeID int

	if err = tx.QueryRow(`
	SELECT task_id, task_type_id, task_data FROM tasks
		WHERE task_id = (
		SELECT task_id FROM pending_queue
			WHERE item_priority > 0             -- 未被冻结
				AND item_remaining_attempts > 0 -- 未在处理
			ORDER BY item_priority DESC, task_id ASC
			LIMIT 1
		)`).Scan(&id, &taskTypeID, &taskData); err == sql.ErrNoRows {
		return 0, "", nil, taskqueue.ErrNoTasksInPending
	} else if err != nil {
		return 0, "", nil, err
	}
	taskType = q.getTypeName(taskTypeID)

	if _, err = tx.Exec(`
	UPDATE pending_queue SET item_remaining_attempts = -item_remaining_attempts
		WHERE task_id = ?`, id); err != nil {
		return 0, "", nil, err
	}
	return id, taskType, taskData, nil
}

// ReportProcessResult 是在完成处理取出的任务后, 向队列报告结果的方法
func (q *Queue) ReportProcessResult(id int, result taskqueue.ProcessResult) (err error) {
	q.dbLock.Lock()
	defer q.dbLock.Unlock()

	return q.reportProcessResult(id, result)
}

func (q *Queue) reportProcessResult(id int, result taskqueue.ProcessResult) (err error) {

	var tx *sql.Tx
	if tx, err = q.db.Begin(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	switch result {
	case taskqueue.ProcessResultFailed: //< 计本次尝试
		var remains int
		if err = tx.QueryRow(`SELECT item_remaining_attempts FROM pending_queue WHERE task_id = ?`, id).Scan(&remains); err != nil {
			return err
		}
		remains = -remains // 根据规定, 任务处理时此值为原本值的相反数
		remains--
		if remains > 0 {
			if _, err = tx.Exec(`UPDATE pending_queue SET item_remaining_attempts = ? WHERE task_id = ?`, remains, id); err != nil {
				return err
			}
			break
		} else if remains < 0 {
			panic("?")
		}
		result = taskqueue.ProcessResultGivenUp
		fallthrough
	case // 以下三类都要将任务移出等待队列
		taskqueue.ProcessResultSuccessful,
		taskqueue.ProcessResultGivenUp,
		taskqueue.ProcessResultShouldBeExcluded:
		if _, err = tx.Exec(`DELETE FROM pending_queue WHERE task_id = ?`, id); err != nil {
			return err
		}
		var status taskqueue.TaskStatus
		switch result {
		case taskqueue.ProcessResultSuccessful:
			status = taskqueue.TaskStatusFinished
		case taskqueue.ProcessResultGivenUp:
			status = taskqueue.TaskStatusGivenUp
		case taskqueue.ProcessResultShouldBeExcluded:
			status = taskqueue.TaskStatusExcluded
		default:
			panic("?")
		}
		if _, err = tx.Exec(`UPDATE tasks SET task_status = ? WHERE task_id = ?`, status, id); err != nil {
			return err
		}
	case taskqueue.ProcessResultRetry: //< 不计本次尝试
		if _, err = tx.Exec(`UPDATE pending_queue SET item_remaining_attempts = -item_remaining_attempts WHERE task_id = ?`, id); err != nil {
			return err
		}
	case taskqueue.ProcessResultShouldBeFrozen: //< 冻结任务
		if _, err = tx.Exec(`
		UPDATE pending_queue
			SET item_priority = -item_priority, -- 其为负值时任务认为是被冻结了的
			item_remaining_attempts = -item_remaining_attempts
			WHERE task_id = ?`, id); err != nil {
			return err
		}
	default:
		panic("?")
	}
	return nil
}
