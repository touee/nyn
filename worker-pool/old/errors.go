package workerpool_old

import "errors"

// ErrInvaildWorkerCapacity 是设置了不正确的 worker 容量时返回的错误
var ErrInvaildWorkerCapacity = errors.New("nyn: worker capacity must > 0")
