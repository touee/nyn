package main

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/touee/nyn"
	components "github.com/touee/nyn/common-components"
	"github.com/touee/nyn/logger"
	taskqueue "github.com/touee/nyn/task-queue"
)

// SimpleTask 是一个爬虫任务
type SimpleTask struct {
	// 抓取 GetURL() 所返回的 URL
	*components.HTTPURLGetFetcher
	// 自动将抓取到的 http.Request.Body 转换为 goquery.Document
	*components.HTTPResponseGoqueryDecorator

	URL string
}

// GetURL 是任务的 URL
func (t SimpleTask) GetURL() string {
	return t.URL
}

// Filter 过滤任务
func (t SimpleTask) Filter(c *nyn.Crawler, _ nyn.Task) (nyn.FilterResult, error) {
	var lock = c.Global["count-lock"].(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	c.Global["count"] = c.Global["count"].(int) + 1
	if c.Global["count"].(int) > 100 {
		return nyn.FilterResultShouldBeFrozen, nil
	}
	return nyn.FilterResultPass, nil
}

// Process 处理任务
func (t SimpleTask) Process(c *nyn.Crawler, _ nyn.Task, payload interface{}) (result taskqueue.ProcessResult, err error) {
	c.TaskLog(logger.LInfo, "processing task", logger.Fields{{"url", t.URL}})

	var taskURL *url.URL
	if taskURL, err = url.Parse(t.URL); err != nil {
		panic(err)
	}

	payload.(*goquery.Document).Find(`a[href]`).Each(func(_ int, sel *goquery.Selection) {
		if href, ok := sel.Attr(`href`); !ok {
			panic(href)
		} else {
			var hrefURL *url.URL
			if hrefURL, err = url.Parse(href); err != nil {
				panic(err)
			}
			hrefURL = taskURL.ResolveReference(hrefURL)
			hrefURL.RawQuery = ""
			hrefURL.Fragment = ""
			if err = c.Request(SimpleTask{URL: hrefURL.String()}); err != nil {
				panic(err)
			}
		}
	})

	return taskqueue.ProcessResultSuccessful, nil
}

func main() {
	var err error

	var dir = fmt.Sprintf("output-simple-%s/crawling", time.Now().Format("20060102150405"))
	if err = os.MkdirAll(dir, 0644); err != nil {
		panic(err)
	}

	var c *nyn.Crawler
	if c, err = nyn.NewCrawler(nyn.CrawlerOptions{
		Dir:                       dir,
		DefaultWorkerPoolWorkers:  20,
		DefaultFileLoggerLogLevel: logger.LTrace,
		//DefaultStdoutLoggerLogLevel: logger.LTrace,
	}); err != nil {
		panic(err)
	}

	if err = c.RegisterTaskTypes(SimpleTask{}); err != nil {
		panic(err)
	}

	c.Global["count"] = 0
	c.Global["count-lock"] = &sync.Mutex{}

	// ¯\_(ツ)_/¯
	if err = c.Request(SimpleTask{URL: "http://go-colly.org/"}); err != nil {
		panic(err)
	}

	c.Run()
	if errs := c.WaitQuit(); errs != nil {
		panic(errs)
	}
}
