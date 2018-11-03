package s3queue_test

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	taskqueue "github.com/touee/nyn/task-queue"
	s3queue "github.com/touee/nyn/task-queue/sqlite3"
)

type DequeueTable struct {
	expectedTaskType string
	expectedTaskData []byte
	expectedTaskID   int
	expectedErr      error
}

func TestDequeue(t *testing.T) {
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

	for _, table := range []EnqueueTable{
		{"a", []byte("1"), taskqueue.EnqueueOptions{Priority: 1}, 2, nil},
		{"b", []byte("1"), taskqueue.EnqueueOptions{Priority: 1, Frozen: true}, 4, nil},
		{"a", []byte("1"), taskqueue.EnqueueOptions{Priority: 1}, 2, nil},
		{"b", []byte{}, taskqueue.EnqueueOptions{Priority: 1, ToHead: true}, -5, nil},
		{"a", []byte{}, taskqueue.EnqueueOptions{Priority: 2}, 6, nil},
		{"c", []byte("123"), taskqueue.EnqueueOptions{Priority: 2, ToHead: true}, -8, nil},
		{"c", []byte("456"), taskqueue.EnqueueOptions{Priority: 1}, 9, nil},
	} {
		testEnqueue(t, q, table)
	}

	for _, table := range []DequeueTable{
		{"c", []byte("123"), -8, nil},
		{"a", []byte{}, 6, nil},
		{"b", []byte{}, -5, nil},
		{"a", []byte("1"), 2, nil},
		{"c", []byte("456"), 9, nil},
		{"", nil, 0, taskqueue.ErrNoTasksInPending},
	} {
		testDequeue(t, q, table)
	}
}

func testDequeue(t *testing.T, q *s3queue.Queue, table DequeueTable) {
	var id, taskType, taskData, err = q.DequeueForProcess()
	if err != table.expectedErr {
		t.Fatalf("expected error: %v, got error: %v", table.expectedErr, err)
	}

	if id != table.expectedTaskID || taskType != table.expectedTaskType || bytes.Compare(taskData, table.expectedTaskData) != 0 {
		t.Fatalf("expected: %#v, got %#v", table, DequeueTable{taskType, taskData, id, err})
	}
}

func TestReportProcessResultA(t *testing.T) {
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

	// frozen

	if _, err = q.EnqueueWithOptions("a", []byte("toBeFrozen"), taskqueue.EnqueueOptions{Priority: 1}); err != nil {
		t.Fatal(err)
	}

	var id int
	if id, _, _, err = q.DequeueForProcess(); err != nil {
		t.Fatal(err)
	}
	if err = q.ReportProcessResult(id, taskqueue.ProcessResultShouldBeFrozen); err != nil {
		t.Fatal(err)
	}

	var frozenID = -1
	q.BatchRefreshStatuses(taskqueue.TaskStatusSet(taskqueue.TaskStatusFrozen), nil, func(status taskqueue.TaskStatus, id int, _ string, _ []byte) taskqueue.TaskStatus {
		frozenID = id
		return taskqueue.TaskStatusPending
	})
	if frozenID != id {
		t.Fatal(frozenID)
	}

	// successful

	if _, err = q.EnqueueWithOptions("a", []byte("toBeSucceed"), taskqueue.EnqueueOptions{Priority: 1}); err != nil {
		t.Fatal(err)
	}

	for {
		if _, _, _, err = q.DequeueForProcess(); err == taskqueue.ErrNoTasksInPending {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if err = q.ReportProcessResult(id, taskqueue.ProcessResultSuccessful); err != nil {
			t.Fatal(err)
		}
	}

	var pendingQueueTaskCount int
	q.BatchRefreshStatuses(taskqueue.TaskStatusSet(taskqueue.TaskStatusPending|taskqueue.TaskStatusFrozen), nil, func(status taskqueue.TaskStatus, id int, _ string, _ []byte) taskqueue.TaskStatus {
		pendingQueueTaskCount++
		return status
	})
	if pendingQueueTaskCount != 0 {
		t.Fatal(pendingQueueTaskCount)
	}
}

func TestReportProcessResultB(t *testing.T) {
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

	const defaultMaxAttempts = 7
	q.SetDefaultMaxAttempts(defaultMaxAttempts)

	type ReportProcessResultTableB struct {
		MaxAttempts  int
		GiveUpInHalf bool
	}

	for i, table := range []ReportProcessResultTableB{
		{00, false},
		{10, false},
		{10, true},
	} {
		var iStr = strconv.Itoa(i)
		var expectedAttemptTimes = table.MaxAttempts
		if expectedAttemptTimes == 0 {
			expectedAttemptTimes = defaultMaxAttempts
		}
		if _, err = q.EnqueueWithOptions("a", []byte(iStr), taskqueue.EnqueueOptions{Priority: 1, MaxAttempts: table.MaxAttempts}); err != nil {
			t.Fatal(err)
		}

		for j := 0; j < expectedAttemptTimes; j++ {
			var id int
			if id, _, _, err = q.DequeueForProcess(); err != nil {
				t.Fatal(err)
			} else if id != i+2 {
				t.Fatal(id)
			}
			for k := 0; k < expectedAttemptTimes*2; k++ {
				if err = q.ReportProcessResult(id, taskqueue.ProcessResultRetry); err != nil {
					t.Fatal(err)
				}
			}
			if j == expectedAttemptTimes/2 && table.GiveUpInHalf {
				if err = q.ReportProcessResult(id, taskqueue.ProcessResultGivenUp); err != nil {
					t.Fatal(err)
				}
				break
			} else {
				if err = q.ReportProcessResult(id, taskqueue.ProcessResultFailed); err != nil {
					t.Fatal(err)
				}
			}
		}
		if _, _, _, err = q.DequeueForProcess(); err != taskqueue.ErrNoTasksInPending {
			t.Fatal(err)
		}
	}
}
