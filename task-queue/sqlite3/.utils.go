package s3queue

import (
	"strconv"
	"strings"

	taskqueue "github.com/touee/nyn/task-queue"
)

func buildSQLListFromStatusSet(set taskqueue.TaskStatusSet) (fragment string) {

	fragment = "("

	for _, status := range set.GetStatuses() {
		fragment += strconv.Itoa(int(status)) + ", "
	}

	fragment = strings.TrimSuffix(fragment, ", ") + ")"

	return fragment
}
