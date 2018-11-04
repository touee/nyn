package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime/pprof"
	"sync"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/touee/nyn"
	"github.com/touee/nyn/logger"
)

func main() {
	var err error

	{
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	var c *nyn.Crawler

	var dir = fmt.Sprintf("output-kaifulee-blog-%s", time.Now().Format("20060102150405"))
	os.Mkdir(dir, 0644)
	var crawlerDir = path.Join(dir, "crawling")
	os.Mkdir(crawlerDir, 0644)

	if c, err = nyn.NewCrawler(nyn.CrawlerOptions{
		Dir:                         crawlerDir,
		DefaultWorkerPoolWorkers:    20,
		DefaultFileLoggerLogLevel:   logger.LTrace,
		DefaultStdoutLoggerLogLevel: logger.LTrace,
		//NoDefaultStdoutLogger:     true,
		//NoDefaultFileLogger: true,
	}); err != nil {
		panic(err)
	}

	var db *sql.DB
	if db, err = sql.Open("sqlite3", path.Join(dir, "result.s3db")); err != nil {
		panic(err)
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS kaifulee_blog (
		title            TEXT    NOT NULL,
		url              TEXT    PRIMARY KEY,
		publication_time INTEGER NOT NULL,
		content          TEXT    NOT NULL,
		tags             TEXT    NOT NULL,
		category         TEXT    NOT NULL,
		read_count       INTEGER NULL,
		comment_count    INTEGER NULL
	)`); err != nil {
		panic(err)
	}
	c.Global["db"] = db
	c.Global["db-lock"] = &sync.Mutex{}
	defer db.Close()
	if c.Global["location"], err = time.LoadLocation("Asia/Shanghai"); err != nil {
		panic(err)
	}

	c.Global["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Safari/605.1.15"

	if err = c.RegisterTaskTypes(ArticleListTask{}, ArticleTask{}, ArticleStatusesTask{}); err != nil {
		panic(err)
	}

	if err = c.Request(ArticleListTask{
		ID:         1197161814,
		PageNumber: 1,
	}); err != nil {
		panic(err)
	}

	c.Run()

	var errs = c.WaitQuit()
	if errs != nil {
		fmt.Printf("%#v", errs)
	}

}
