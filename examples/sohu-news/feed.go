package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/mattn/go-sqlite3"

	"github.com/touee/nyn"
	components "github.com/touee/nyn/common-components"
	"github.com/touee/nyn/logger"
	taskqueue "github.com/touee/nyn/task-queue"
)

// FeedTask 是获取 feed 的任务
type FeedTask struct {
	components.HTTPURLGetFetcher
	components.HTTPResponseJSONDecorator

	SceneID int
	Page    int
}

// Feeds 表示 feed api 返回的 json 的结构
type Feeds []Article

// Article 表示 feed api 返回的 json 中, 单篇文章的结构
type Article struct {
	ID                   int          `json:"id"`
	AuthorID             int          `json:"authorId"`
	AuthorName           string       `json:"authorName"`
	Title                string       `json:"title"`
	Tags                 []ArticleTag `json:"tags"`
	PublicationTimeMilli int64        `json:"publicTime"`
	SourceURL            string       `json:"originalSource"`
}

// ArticleTag 表示 Article 所属的 tag
type ArticleTag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetURL 获取 URL
func (task FeedTask) GetURL() string {
	return fmt.Sprintf("http://v2.sohu.com/public-api/feed?scene=CATEGORY&sceneId=%d&page=%d&size=100", task.SceneID, task.Page)
}

var feedsType = reflect.TypeOf(Feeds{})

// GetPayloadType 是 Process 所要的 payload 的类型
func (task FeedTask) GetPayloadType() reflect.Type {
	return feedsType
}

// Process 处理获取到的 feed
func (task FeedTask) Process(c *nyn.Crawler, _ nyn.Task, payload interface{}) (result taskqueue.ProcessResult, err error) {
	c.TaskLog(logger.LInfo, "开始处理 feed", logger.Fields{{"task", task}})
	var feeds = payload.(Feeds)

	var lock = c.Global["db-lock"].(*sync.Mutex)
	lock.Lock()

	var tx *sql.Tx
	if tx, err = c.Global["db"].(*sql.DB).Begin(); err != nil {
		panic(err)
	}
	for _, feed := range feeds {
		var t = time.Unix(0, feed.PublicationTimeMilli*int64(time.Millisecond)-time.Hour.Nanoseconds()*8)
		var tags string
		if tagsBytes, err := json.Marshal(feed.Tags); err != nil {
			panic(err)
		} else {
			tags = string(tagsBytes)
		}

		var newTask = ArticleTask{ID: feed.ID, AuthorID: feed.AuthorID}

		if _, err = tx.Exec(`
		INSERT INTO sohu_news (title, url, article_id, author_id, author_name, publication_time, tags, source_url) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, feed.Title, newTask.GetURL(), feed.ID, feed.AuthorID, feed.AuthorName, t.Unix(), tags, feed.SourceURL); err != nil {
			if err == sqlite3.ErrConstraintUnique {
				c.TaskLog(logger.LWarning, "重复新闻", logger.Fields{{"newTask", newTask}})
			} else {
				panic(err)
			}
		}

		c.Request(newTask)
	}
	tx.Commit()

	lock.Unlock()

	if task.Page < 100 {
		if err = c.Request(FeedTask{SceneID: task.SceneID, Page: task.Page + 1}); err != nil {
			panic(err)
		}
	}

	return taskqueue.ProcessResultSuccessful, nil
}
