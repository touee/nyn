package main

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/touee/nyn"
	components "github.com/touee/nyn/common-components"
	"github.com/touee/nyn/logger"
	taskqueue "github.com/touee/nyn/task-queue"
)

// ArticleTask 文章任务
type ArticleTask struct {
	components.HTTPURLGetFetcher
	components.HTTPResponseGoqueryDecorator

	ID       int
	AuthorID int
}

// GetURL 获取任务对应的 URL
func (task ArticleTask) GetURL() string {
	return fmt.Sprintf("http://www.sohu.com/a/%d_%d", task.ID, task.AuthorID)
}

// Process 处理获取到的文章
func (task ArticleTask) Process(c *nyn.Crawler, _ nyn.Task, payload interface{}) (result taskqueue.ProcessResult, err error) {
	c.TaskLog(logger.LInfo, "开始处理文章", logger.Fields{{"task", task}})

	var doc = payload.(*goquery.Document)

	var (
		url = task.GetURL()

		content string
	)

	if content, err = doc.Find(".article").Html(); err != nil {
		c.TaskLog(logger.LError, "获取内容失败", logger.Fields{{"url", url}, {"err", err}})
	}

	var lock = c.Global["db-lock"].(*sync.Mutex)
	lock.Lock()
	if _, err = c.Global["db"].(*sql.DB).Exec(`
	UPDATE sohu_news SET content = ? WHERE article_id = ?
	`, content, task.ID); err != nil {
		panic(err)
	}
	lock.Unlock()

	if err = c.Request(ArticleCommentsTask{ID: task.ID, PageNO: 1}, PVTask{ID: task.ID}); err != nil {
		panic(err)
	}

	return taskqueue.ProcessResultSuccessful, err
}
