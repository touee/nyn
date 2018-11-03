package components

import "fmt"

// BadStatusCodeError 是当响应的状态码并非 2xx 时返回的错误
type BadStatusCodeError struct {
	StatusCode int
}

func (err BadStatusCodeError) Error() string {
	return fmt.Sprintf("bad status code: %d", err.StatusCode)
}

// Temporary 返回此错误是否是临时性的
func (err BadStatusCodeError) Temporary() bool {
	if err.StatusCode >= 500 && err.StatusCode <= 599 {
		return true
	}
	return false
}
