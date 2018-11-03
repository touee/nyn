package s3queue

import (
	"database/sql"

	taskqueue "github.com/touee/nyn/task-queue"
)

// EnqueueWithOptions 将任务根据 opts 的设置放入队列
func (q *Queue) EnqueueWithOptions(taskType string, taskData []byte, opts taskqueue.EnqueueOptions) (id int, err error) {
	q.dbLock.Lock()
	var typeID = q.getTypeID(taskType, true)
	defer q.dbLock.Unlock()
	return q.enqueueWithOptions(typeID, taskData, opts)
}

func (q *Queue) enqueueWithOptions(typeID int, taskData []byte, opts taskqueue.EnqueueOptions) (id int, err error) {
	var taskID int
	var status taskqueue.TaskStatus

	var tx *sql.Tx
	if tx, err = q.db.Begin(); err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if err = tx.QueryRow(`SELECT task_id, task_status FROM tasks WHERE task_type_id = ? AND task_data = ?`, typeID, taskData).Scan(&taskID, &status); err == sql.ErrNoRows { //< 任务不存在
		taskID = q.nextI()

		if opts.ToHead {
			taskID = -taskID
		}

		if _, err = tx.Exec(`INSERT INTO tasks (task_id, task_type_id, task_data, task_status) VALUES (?, ?, ?, ?)`, taskID, typeID, taskData, taskqueue.TaskStatusPending /* tasks 表中, Pending/Frozen/Processing 都归为 Pending */); err != nil {
			return 0, err
		}

		var priority = opts.Priority
		if priority <= 0 {
			return 0, taskqueue.ErrInvaildPriority
		} else if opts.Frozen {
			priority = -priority
		}
		var remainAttempts = opts.MaxAttempts
		if opts.MaxAttempts == 0 {
			remainAttempts = q.defaultMaxAttempts
		}

		if _, err = tx.Exec(`INSERT INTO pending_queue (task_id, item_priority, item_remaining_attempts) VALUES (?, ?, ?)`, taskID, priority, remainAttempts); err != nil {
			return 0, err
		}
		return taskID, nil
	} else if err != nil { //< 有错误
		return 0, err
	} else { // 任务本身存在
		switch status {
		case taskqueue.TaskStatusPending:
			// TODO
		case taskqueue.TaskStatusGivenUp:
			// TODO
		}
		return taskID, nil
	}
}
