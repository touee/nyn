package taskqueue

// TaskStatus 是任务在 task queue 中的状态
// 小于等于 0 的值代表已尝试的次数
type TaskStatus uint8

const (
	// TaskStatusPending 代表任务正在等待执行
	TaskStatusPending TaskStatus = 1
	// TaskStatusFrozen 代表任务已被冻结, 在被解冻之前不会被取出队列
	TaskStatusFrozen TaskStatus = 1 << 1

	// TaskStatusProcessing 代表任务正在被处理
	TaskStatusProcessing TaskStatus = 1 << 2

	// TaskStatusGivenUp 代表超过尝试次数而放弃的任务
	TaskStatusGivenUp TaskStatus = 1 << 3

	// TaskStatusExcluded 代表任务被除外
	TaskStatusExcluded TaskStatus = 1 << 6
	// TaskStatusFinished 代表任务已完成
	TaskStatusFinished TaskStatus = 1 << 7
)

func (s TaskStatus) String() (name string) {
	name = map[TaskStatus]string{
		TaskStatusPending:    "Pending",
		TaskStatusProcessing: "Processing",
		TaskStatusFrozen:     "Frozen",
		TaskStatusGivenUp:    "GivenUp",
		TaskStatusFinished:   "Finished",
	}[s]
	if name == "" {
		return "invalid"
	}
	return name
}

// TaskStatusSet 是任务状态的集合
type TaskStatusSet TaskStatus

// GetStatuses 返回 TaskStatusSet 中的所有状态
func (set TaskStatusSet) GetStatuses() (ss []TaskStatus) {
	for i := TaskStatus(1); i != 0; i <<= 1 {
		if uint8(set)&uint8(i) != 0 {
			ss = append(ss, i)
		}
	}
	return ss
}

/*
// ContainsAny 返回 TaskStatusSet 是否包含多个所给状态中的至少一个状态某一状态
func (set TaskStatusSet) ContainsAny(statuses ...TaskStatus) bool {
	var set2 TaskStatusSet
	for _, status := range statuses {
		set2 |= TaskStatusSet(status)
	}
	return set&set2 != 0
}
*/
