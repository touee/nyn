package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/touee/nyn"
	components "github.com/touee/nyn/common-components"
	"github.com/touee/nyn/logger"
	taskqueue "github.com/touee/nyn/task-queue"
)

// PVTask 是获取文章阅读量的任务
type PVTask struct {
	components.HTTPURLGetFetcher

	ID int
}

// GetURL 获取任务对应的 URL
func (task PVTask) GetURL() string {
	return fmt.Sprintf("http://v2.sohu.com/public-api/articles/%d/pv", task.ID)
}

// Process 处理获取到的文章阅读量
func (task PVTask) Process(c *nyn.Crawler, _ nyn.Task, payload interface{}) (result taskqueue.ProcessResult, err error) {
	c.TaskLog(logger.LInfo, "开始处理文章阅读量", logger.Fields{{"task", task}})

	var resp = payload.(*http.Response)
	defer resp.Body.Close()

	var rawCount []byte
	if rawCount, err = ioutil.ReadAll(resp.Body); err != nil {
		panic(err)
	}

	var count int
	if count, err = strconv.Atoi(string(rawCount)); err != nil {
		panic(err)
	}

	var lock = c.Global["db-lock"].(*sync.Mutex)
	lock.Lock()
	if _, err = c.Global["db"].(*sql.DB).Exec(`
	UPDATE sohu_news SET read_count = ? WHERE article_id = ?
	`, count, task.ID); err != nil {
		panic(err)
	}
	lock.Unlock()

	return taskqueue.ProcessResultSuccessful, err

}
