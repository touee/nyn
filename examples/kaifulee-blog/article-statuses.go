package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"

	"github.com/touee/nyn"
	components "github.com/touee/nyn/common-components"
	"github.com/touee/nyn/logger"
	taskqueue "github.com/touee/nyn/task-queue"
)

// ArticleStatusesTask 文章状态任务
type ArticleStatusesTask struct {
	components.HTTPURLGetFetcher

	UID string
	AID string
}

// GetURL 获取 URL
func (task ArticleStatusesTask) GetURL() string {
	// http://comet.blog.sina.com.cn/api?maintype=num&uid=475b3d56&aids=02vo2k
	return fmt.Sprintf("http://comet.blog.sina.com.cn/api?maintype=num&uid=%s&aids=%s", task.UID, task.AID)
}

// $ScriptLoader.response("",{"02vo2k":{"f":16,"d":794,"r":63709,"c":81,"z":153}})
var statusesRX = regexp.MustCompile(`:(\{.*?\})`)

// Process 处理获取到的文章状态
func (task ArticleStatusesTask) Process(c *nyn.Crawler, _ nyn.Task, payload interface{}) (result taskqueue.ProcessResult, err error) {
	c.TaskLog(logger.LTrace, "开始处理文章状态", logger.Fields{{"task", task}})

	var resp = payload.(*http.Response)
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		panic(err)
	} else {
		if submatches := statusesRX.FindSubmatch(body); len(submatches) != 2 {
			c.TaskLog(logger.LFatal, "无法正则匹配文章列表 URL", logger.Fields{{"submatchs", submatches}, {"raw", string(body)}})
			panic(string(body))
		} else {
			var statuses struct {
				ReadCount    int `json:"r"`
				CommentCount int `json:"c"`
			}
			if err = json.Unmarshal(submatches[1], &statuses); err != nil {
				panic(err)
			}

			var lock = c.Global["db-lock"].(*sync.Mutex)
			lock.Lock()
			if _, err = c.Global["db"].(*sql.DB).Exec(`
			UPDATE kaifulee_blog SET read_count = ?, comment_count = ? WHERE url = ?
			`, statuses.ReadCount, statuses.CommentCount, ArticleTask{UID: task.UID, AID: task.AID}.GetURL()); err != nil {
				panic(err)
			}
			lock.Unlock()

			return taskqueue.ProcessResultSuccessful, nil
		}
	}
}
