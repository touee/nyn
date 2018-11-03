package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/touee/nyn"
	components "github.com/touee/nyn/common-components"
	"github.com/touee/nyn/logger"
	taskqueue "github.com/touee/nyn/task-queue"
)

// ArticleListTask 文章列表任务
type ArticleListTask struct {
	components.HTTPURLGetFetcher
	components.HTTPResponseGoqueryDecorator

	ID         int
	PageNumber int
}

// GetURL 获取 URL
func (task ArticleListTask) GetURL() string {
	return fmt.Sprintf("http://blog.sina.com.cn/s/articlelist_%d_0_%d.html", task.ID, task.PageNumber)
}

// http://blog.sina.com.cn/s/blog_475b3d560102y8ke.html
var articleURLRX = regexp.MustCompile(`blog_(.{8})01(.{6})\.html`)

// http://blog.sina.com.cn/s/articlelist_1197161814_0_1.html
var articleListURLRX = regexp.MustCompile(`articlelist_(.*)_0_(.*)\.html`)

// Process 处理获取到的文章列表
func (task ArticleListTask) Process(c *nyn.Crawler, _ nyn.Task, payload interface{}) (result taskqueue.ProcessResult, err error) {
	var doc = payload.(*goquery.Document)

	c.TaskLog(logger.LTrace, "开始处理文章列表", logger.Fields{{"task", task}})

	//html, _ := doc.Html()
	//ioutil.WriteFile("test.html", []byte(html), 0644)

	doc.Find(`.atc_title a`).Each(func(_ int, s *goquery.Selection) {
		if href, ok := s.Attr(`href`); ok {
			if submatches := articleURLRX.FindStringSubmatch(href); len(submatches) != 3 {
				c.TaskLog(logger.LFatal, "无法正则匹配文章列表 URL", logger.Fields{{"submatchs", submatches}, {"raw", href}})
				panic(href)
			} else {
				c.Request(ArticleTask{UID: submatches[1], AID: submatches[2]})
			}
		}
	})

	doc.Find(`.SG_pages a`).Each(func(_ int, s *goquery.Selection) {
		if href, ok := s.Attr(`href`); ok {
			if submatchs := articleListURLRX.FindStringSubmatch(href); len(submatchs) != 3 {
				c.TaskLog(logger.LFatal, "无法正则匹配文章列表 URL", logger.Fields{{"submatchs", submatchs}, {"raw", href}})
				panic(href)
			} else {
				var id, pn int
				var rawID, rawPN = submatchs[1], submatchs[2]
				if id, err = strconv.Atoi(rawID); err != nil {
					panic(rawID)
				}
				if pn, err = strconv.Atoi(rawPN); err != nil {
					panic(rawPN)
				}

				c.Request(ArticleListTask{ID: id, PageNumber: pn})
			}
		}
	})

	return taskqueue.ProcessResultSuccessful, nil
}
