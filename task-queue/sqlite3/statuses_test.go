package s3queue_test

import (
	"os"
	"testing"

	taskqueue "github.com/touee/nyn/task-queue"
	s3queue "github.com/touee/nyn/task-queue/sqlite3"
)

func TestBatchUpdateStatuses(t *testing.T) {
	var tempFileName = getTempFileName(t.Name())
	defer func() {
		os.Remove(tempFileName)
	}()

	var (
		err error
		q   *s3queue.Queue
	)
	q, err = s3queue.Open(tempFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = q.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// no. 1

	for _, table := range []EnqueueTable{
		{"a", []byte{2}, taskqueue.EnqueueOptions{Priority: 1, Frozen: true}, 2, nil},
		{"a", []byte{3}, taskqueue.EnqueueOptions{Priority: 1}, 3, nil},
		{"a", []byte{4}, taskqueue.EnqueueOptions{Priority: 1}, 4, nil},
		{"a", []byte{5}, taskqueue.EnqueueOptions{Priority: 1, Frozen: true}, 5, nil},
	} {
		testEnqueue(t, q, table)
	}

	for _, table := range []DequeueTable{
		DequeueTable{"a", []byte{3}, 3, nil},
		DequeueTable{"a", []byte{4}, 4, nil},
		DequeueTable{"", nil, 0, taskqueue.ErrNoTasksInPending},
	} {
		testDequeue(t, q, table)
	}

	// no. 2

	for _, table := range []EnqueueTable{
		{"a", []byte{6}, taskqueue.EnqueueOptions{Priority: 1}, 6, nil},
		{"a", []byte{7}, taskqueue.EnqueueOptions{Priority: 1, Frozen: true}, 7, nil},
	} {
		testEnqueue(t, q, table)
	}

	q.BatchRefreshStatuses(taskqueue.TaskStatusSet(taskqueue.TaskStatusFrozen), nil, func(status taskqueue.TaskStatus, id int, taskType string, taskData []byte) taskqueue.TaskStatus {
		if id == 5 {
			return taskqueue.TaskStatusPending
		} else if id == 6 { // 不会生效
			return taskqueue.TaskStatusFrozen
		}
		return status
	})

	for _, table := range []DequeueTable{
		DequeueTable{"a", []byte{5}, 5, nil},
		DequeueTable{"a", []byte{6}, 6, nil},
		DequeueTable{"", nil, 0, taskqueue.ErrNoTasksInPending},
	} {
		testDequeue(t, q, table)
	}

	// no. 3

	for _, table := range []EnqueueTable{
		{"a", []byte{8}, taskqueue.EnqueueOptions{Priority: 1}, 8, nil},
		{"a", []byte{9}, taskqueue.EnqueueOptions{Priority: 1, Frozen: true}, 9, nil},
	} {
		testEnqueue(t, q, table)
	}

	q.BatchRefreshStatuses(taskqueue.TaskStatusSet(taskqueue.TaskStatusPending), nil, func(status taskqueue.TaskStatus, id int, taskType string, taskData []byte) taskqueue.TaskStatus {
		if id == 8 {
			return taskqueue.TaskStatusFrozen
		} else if id == 9 { // 不会生效
			return taskqueue.TaskStatusPending
		}
		return status
	})

	for _, table := range []DequeueTable{
		DequeueTable{"", nil, 0, taskqueue.ErrNoTasksInPending},
	} {
		testDequeue(t, q, table)
	}

	// no. 4

	for _, table := range []EnqueueTable{
		{"a", []byte{10}, taskqueue.EnqueueOptions{Priority: 1}, 10, nil},
		{"a", []byte{11}, taskqueue.EnqueueOptions{Priority: 1, Frozen: true}, 11, nil},
		{"a", []byte{12}, taskqueue.EnqueueOptions{Priority: 1}, 12, nil},
	} {
		testEnqueue(t, q, table)
	}

	q.BatchRefreshStatuses(taskqueue.TaskStatusSet(taskqueue.TaskStatusPending|taskqueue.TaskStatusFrozen), nil, func(status taskqueue.TaskStatus, id int, taskType string, taskData []byte) taskqueue.TaskStatus {
		if status == taskqueue.TaskStatusPending {
			return taskqueue.TaskStatusFrozen
		} else {
			return taskqueue.TaskStatusPending
		}
	})

	for _, table := range []DequeueTable{
		DequeueTable{"a", []byte{2}, 2, nil},
		DequeueTable{"a", []byte{7}, 7, nil},
		DequeueTable{"a", []byte{8}, 8, nil},
		DequeueTable{"a", []byte{9}, 9, nil},
		DequeueTable{"a", []byte{11}, 11, nil},
		DequeueTable{"", nil, 0, taskqueue.ErrNoTasksInPending},
	} {
		testDequeue(t, q, table)
	}

}
