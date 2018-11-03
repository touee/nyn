package taskqueue_test

import (
	"testing"

	taskqueue "github.com/touee/nyn/task-queue"
)

func TestTaskStatusSetGettingStatuses(t *testing.T) {
	var tables = []struct {
		set            taskqueue.TaskStatusSet
		resultStatuses []taskqueue.TaskStatus
	}{
		{
			taskqueue.TaskStatusSet(taskqueue.TaskStatusFinished | taskqueue.TaskStatusFrozen),
			[]taskqueue.TaskStatus{taskqueue.TaskStatusFinished, taskqueue.TaskStatusFrozen},
		},
	}
	for _, table := range tables {
		var statuses = table.set.GetStatuses()
		if len(statuses) != len(table.resultStatuses) {
			t.Fatal()
		}
		var set = make(map[taskqueue.TaskStatus]struct{})
		for _, status := range statuses {
			set[status] = struct{}{}
		}
		for _, status := range table.resultStatuses {
			delete(set, status)
		}
		if len(set) != 0 {
			t.Fatal()
		}
	}
}
