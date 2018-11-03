package nyn_test

import (
	"testing"

	"github.com/touee/nyn/logger"

	"github.com/touee/nyn"
)

// TestNewCrawler 新建打开爬虫
func TestNewCrawler(t *testing.T) {
	var c, err = nyn.NewCrawler(nyn.CrawlerOptions{
		DefaultStdoutLoggerLogLevel: logger.LDebug,
		NoDefaultFileLogger:         true,
	})
	if err != nil {
		t.Fatal(err)
	}

	_ = c
}
