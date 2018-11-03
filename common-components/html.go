package components

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/touee/nyn"
)

// HTTPResponseGoqueryDecorator 将得到的 HTTP 响应中的 body 转换为 goquery 文档
// 响应的 body 应是 html/xml 文档
type HTTPResponseGoqueryDecorator struct{}

// DecoratePayload 转换 payload
func (HTTPResponseGoqueryDecorator) DecoratePayload(c *nyn.Crawler, _ nyn.Task, payload interface{}) (decoratedPayload interface{}, err error) {
	var resp = payload.(*http.Response)
	defer resp.Body.Close()

	//var doc *goquery.Document
	return goquery.NewDocumentFromReader(resp.Body)
}
