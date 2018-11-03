package s3queue

import (
	"database/sql"
	"fmt"

	taskqueue "github.com/touee/nyn/task-queue"
)

// BatchRefreshStatuses 批量刷新任务的状态
func (q *Queue) BatchRefreshStatuses(appliedStatuses taskqueue.TaskStatusSet, appliedTaskTypes []string, filter taskqueue.Refresher) (err error) {
	q.dbLock.Lock()
	defer q.dbLock.Unlock()
	return q.batchRefreshStatuses(appliedStatuses, appliedTaskTypes, filter)
}

func (q *Queue) batchRefreshStatuses(appliedStatuses taskqueue.TaskStatusSet, appliedTaskTypes []string, filter taskqueue.Refresher) (err error) {
	if appliedStatuses == 0 {
		return nil
	} else if appliedStatuses&
		taskqueue.TaskStatusSet(taskqueue.TaskStatusProcessing) != 0 {
		panic("You can't")
	} else if appliedStatuses&
		taskqueue.TaskStatusSet(
			taskqueue.TaskStatusGivenUp|
				taskqueue.TaskStatusExcluded|
				taskqueue.TaskStatusFinished) != 0 {
		panic("Hasn't implemented")
	}

	if appliedTaskTypes != nil {
		panic("Hasn't implemented")
	}

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

	var pendingApplied = appliedStatuses&taskqueue.TaskStatusSet(taskqueue.TaskStatusPending) != 0
	var frozenApplied = appliedStatuses&taskqueue.TaskStatusSet(taskqueue.TaskStatusFrozen) != 0

	if err = q.batchRefreshStatusesInPendingQueue(tx, pendingApplied, frozenApplied, filter); err != nil {
		return err
	}

	return nil
}

func (q *Queue) batchRefreshStatusesInPendingQueue(tx *sql.Tx, appliesToPendingItems, appliesToFrozenItems bool, filter taskqueue.Refresher) (err error) {
	var whereCondition = `item_remaining_attempts > 0`
	switch {
	case appliesToPendingItems && appliesToFrozenItems:
	case appliesToPendingItems:
		whereCondition += ` AND item_priority > 0`
	case appliesToFrozenItems:
		whereCondition += ` AND item_priority < 0`
	default:
		whereCondition = ``
	}

	if whereCondition == `` {
		return nil
	}

	var rows *sql.Rows
	if rows, err = tx.Query(fmt.Sprintf(`SELECT tasks.task_id, task_type_id, task_data, item_priority FROM tasks JOIN pending_queue WHERE tasks.task_id = pending_queue.task_id AND %s`, whereCondition)); err != nil {
		return err
	}
	defer func() {
		rows.Close()
	}()

	for rows.Next() {
		var (
			id       int
			typeID   int
			data     []byte
			priority int
		)
		if err = rows.Scan(&id, &typeID, &data, &priority); err != nil {
			return err
		}
		var status = taskqueue.TaskStatusPending
		if priority < 0 {
			status = taskqueue.TaskStatusFrozen
		}
		if newStatus := filter(status, id, q.getTypeName(typeID), data); newStatus != status {
			if newStatus == taskqueue.TaskStatusProcessing {
				panic("You can't")
			} else if newStatus != taskqueue.TaskStatusPending &&
				newStatus != taskqueue.TaskStatusFrozen {
				panic("Hasn't implemented")
			}
			if _, err = tx.Exec(`UPDATE pending_queue SET item_priority = -item_priority WHERE task_id = ?`, id); err != nil {
				return err
			}
		}
	}
	return nil
}
