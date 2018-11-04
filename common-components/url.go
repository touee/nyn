package components

import (
	"net/http"
	"time"

	"github.com/touee/nyn"
)

// HTTPURLGetFetcher 包装了 http url get 任务的 fetcher
type HTTPURLGetFetcher struct {
	//URL string
}

var client = http.Client{Timeout: time.Second * 10}

// Fetch 获取内容
// 响应相关的问题/错误应该在这里处理
func (HTTPURLGetFetcher) Fetch(c *nyn.Crawler, task nyn.Task) (payload interface{}, err error) {

	var (
		req  *http.Request
		resp *http.Response
	)

	if req, err = http.NewRequest("GET", task.(interface{ GetURL() string }).GetURL(), nil); err != nil {
		panic(err) // ?
	}
	if ua, ok := c.Global["User-Agent"]; ok {
		req.Header.Set("User-Agent", ua.(string))
	}
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, BadStatusCodeError{resp.StatusCode}
	}
	return resp, nil
}
