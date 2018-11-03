package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

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

	UID string
	AID string
}

// GetURL 获取 URL
func (task ArticleTask) GetURL() string {
	return fmt.Sprintf("http://blog.sina.com.cn/s/blog_%s01%s.html", task.UID, task.AID)
}

// Process 处理获取到的文章
func (task ArticleTask) Process(c *nyn.Crawler, _ nyn.Task, payload interface{}) (result taskqueue.ProcessResult, err error) {
	c.TaskLog(logger.LTrace, "开始处理文章", logger.Fields{{"task", task}})

	var doc = payload.(*goquery.Document)

	var (
		url   = task.GetURL()
		title string

		rawTime         string
		publicationTime time.Time

		content  string
		tags     = make([]string, 0)
		category string
	)

	if main := doc.Find(`.BNE_main`); main.Length() != 0 {

		title = main.Find(`.h1_tit`).Text()

		rawTime = main.Find("#pub_time").Text()

		if content, err = main.Find(".BNE_cont").Html(); err != nil {
			c.TaskLog(logger.LError, "获取内容失败", logger.Fields{{"url", url}, {"err", err}})
		}

	} else {

		main = doc.Find(`#articlebody`)

		title = main.Find(`.titName`).Text()

		rawTime = strings.Trim(main.Find(`.time`).Text(), "()")

		if content, err = main.Find(`.articalContent`).Html(); err != nil {
			c.TaskLog(logger.LError, "获取内容失败", logger.Fields{{"url", url}, {"err", err}})
		}

		main.Find(`.blog_tag`).Find(`h3`).Each(func(_ int, s *goquery.Selection) {
			tags = append(tags, s.Text())
		})

		category = main.Find(`.blog_class a`).Text()

	}

	if publicationTime, err = time.ParseInLocation("2006-01-02 15:04:05", rawTime, c.Global["location"].(*time.Location)); err != nil {
		c.TaskLog(logger.LError, "解析时间失败", logger.Fields{{"url", url}, {"err", err}, {"raw", rawTime}})
	}

	var tagsJSON string
	if _tagsJSON, err := json.Marshal(tags); err != nil {
		panic(err)
	} else {
		tagsJSON = string(_tagsJSON)
	}
	var lock = c.Global["db-lock"].(*sync.Mutex)
	lock.Lock()
	if _, err = c.Global["db"].(*sql.DB).Exec(`
	INSERT INTO kaifulee_blog (title, url, publication_time, content, tags, category)
	VALUES (?, ?, ?, ?, ?, ?)
	`, title, url, publicationTime.Unix(), content, tagsJSON, category); err != nil {
		panic(err)
	}
	lock.Unlock()

	c.Request(ArticleStatusesTask{UID: task.UID, AID: task.AID})

	return taskqueue.ProcessResultSuccessful, nil
}
