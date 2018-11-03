package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	components "github.com/touee/nyn/common-components"
	"github.com/touee/nyn/logger"

	"github.com/touee/nyn"
	taskqueue "github.com/touee/nyn/task-queue"
)

// ArticleCommentsTask 文章任务
type ArticleCommentsTask struct {
	components.HTTPURLGetFetcher
	components.HTTPResponseJSONDecorator

	ID     int
	PageNO int
}

// CommentsBody 是 load api 返回的 json 结构
type CommentsBody struct {
	Code       int    `json:"code"`
	Message    string `json:"msg"`
	JSONObject struct {
		Comments   []Comment `json:"comments"`
		CommentSum int       `json:"cmt_sum"`
	} `json:"jsonObject"`
}

// Comment 是 load api 返回的 json 结构中的单个评论
type Comment struct {
	Comments []struct {
		CommentID int `json:"comment_id"`
	} `json:"comments"`
	CreationTime int64  `json:"create_time"`
	Content      string `json:"content"`
	CommentID    int    `json:"comment_id"`
	Passport     struct {
		UserID   int    `json:"user_id"`
		Nickname string `json:"nickname"`
	} `json:"passport"`
}

// GetURL 获取 URL
func (task ArticleCommentsTask) GetURL() string {
	return fmt.Sprintf("http://apiv2.sohu.com/api/topic/load?page_size=100&page_no=%d&source_id=mp_%d", task.PageNO, task.ID)
}

var commentsType = reflect.TypeOf(CommentsBody{})

// GetPayloadType 是 Process 所要的 payload 的类型
func (task ArticleCommentsTask) GetPayloadType() reflect.Type {
	return commentsType
}

// Process 处理获取到的评论
func (task ArticleCommentsTask) Process(c *nyn.Crawler, _ nyn.Task, payload interface{}) (result taskqueue.ProcessResult, err error) {
	c.TaskLog(logger.LInfo, "开始处理文章评论", logger.Fields{{"task", task}})
	var body = payload.(CommentsBody)
	if body.Code != 200 {
		c.TaskLog(logger.LWarning, "获取评论失败", logger.Fields{{"raw", body}})
		return taskqueue.ProcessResultFailed, nil
	}

	var lock = c.Global["db-lock"].(*sync.Mutex)
	lock.Lock()

	var tx *sql.Tx
	if tx, err = c.Global["db"].(*sql.DB).Begin(); err != nil {
		panic(err)
	}

	if task.PageNO == 1 {
		if _, err = tx.Exec(`
		UPDATE sohu_news SET comment_count = ? WHERE article_id = ?
		`, body.JSONObject.CommentSum, task.ID); err != nil {
			panic(err)
		}
	}

	for _, comment := range body.JSONObject.Comments {
		var referenceIDs = []int{}
		for _, referenceComment := range comment.Comments {
			referenceIDs = append(referenceIDs, referenceComment.CommentID)
		}
		var referenceIDsJSON []byte
		if referenceIDsJSON, err = json.Marshal(referenceIDs); err != nil {
			panic(err)
		}

		if _, err = tx.Exec(`
		INSERT OR IGNORE INTO sohu_users (user_id, nickname) VALUES (?, ?)
		`, comment.Passport.UserID, comment.Passport.Nickname); err != nil {
			panic(err)
		}

		var t = time.Unix(0, comment.CreationTime*int64(time.Millisecond)-time.Hour.Nanoseconds()*8)

		if _, err = tx.Exec(`
		INSERT OR IGNORE INTO sohu_news_comments (article_id, comment_id, creation_time, reference_ids, content, user_id) VALUES (?, ?, ?, ?, ?, ?)
		`, task.ID, comment.CommentID, t, string(referenceIDsJSON), comment.Content, comment.Passport.UserID); err != nil {
			panic(err)
		}

	}

	tx.Commit()

	lock.Unlock()

	if nextPage := task.PageNO + 1; nextPage <= (body.JSONObject.CommentSum-1)/100+1 {
		c.Request(ArticleCommentsTask{ID: task.ID, PageNO: task.PageNO + 1})
	}

	return taskqueue.ProcessResultSuccessful, nil

}
