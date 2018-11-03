package s3queue_test

import (
	"os"
	"testing"

	"github.com/touee/nyn/task-queue"

	s3queue "github.com/touee/nyn/task-queue/sqlite3"
)

type EnqueueTable struct {
	taskType string
	taskData []byte
	opts     taskqueue.EnqueueOptions

	expectedTaskID int
	expectedErr    error
}

func testEnqueue(t *testing.T, q *s3queue.Queue, table EnqueueTable) {
	var err error
	var id int
	if id, err = q.EnqueueWithOptions(table.taskType, table.taskData, table.opts); err != table.expectedErr {
		t.Fatalf("expected error: %v, got error: %v", table.expectedErr, err)
	} else if id != table.expectedTaskID {
		t.Fatalf("expected: %d, got: %d", table.expectedTaskID, id)

	} else {
		t.Log(id)

	}
}

func TestEnqueue(t *testing.T) {
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

	func() {
		defer func() {
			err = q.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		for _, table := range []EnqueueTable{
			{"a", []byte("1"), taskqueue.EnqueueOptions{Priority: 1}, 2, nil},
			{"b", []byte("1"), taskqueue.EnqueueOptions{Priority: 1}, 4, nil},
			{"a", []byte("1"), taskqueue.EnqueueOptions{Priority: 1}, 2, nil},
			{"b", []byte{}, taskqueue.EnqueueOptions{Priority: 1, ToHead: true}, -5, nil},
		} {
			testEnqueue(t, q, table)
		}
	}()

	q, err = s3queue.Open(tempFileName)
	if err != nil {
		t.Fatal(err)
	}

	func() {
		defer func() {
			err = q.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		for _, table := range []EnqueueTable{
			{"a", []byte("2"), taskqueue.EnqueueOptions{Priority: 1}, 6, nil},
			{"a", []byte("1"), taskqueue.EnqueueOptions{Priority: 1}, 2, nil},
			{"b", []byte{}, taskqueue.EnqueueOptions{Priority: 1}, -5, nil},
		} {
			testEnqueue(t, q, table)
		}
	}()

}
