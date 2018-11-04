package main

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/touee/nyn"
)

func main() {
	var err error

	var c *nyn.Crawler

	var dir = fmt.Sprintf("output-sohu-news-%s", time.Now().Format("20060102150405"))
	os.Mkdir(dir, 0644)
	var crawlerDir = path.Join(dir, "crawling")
	os.Mkdir(crawlerDir, 0644)

	if c, err = nyn.NewCrawler(nyn.CrawlerOptions{
		Dir: crawlerDir,
		//NoDefaultStdoutLogger:     true,
		//DefaultFileLoggerLogLevel: logger.LTrace,
		//RemoveDefaultLogger: true,
	}); err != nil {
		panic(err)
	}

	var db *sql.DB
	if db, err = sql.Open("sqlite3", path.Join(dir, "result.s3db")); err != nil {
		panic(err)
	}
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS sohu_news (
			title            TEXT    NOT NULL,
			url              TEXT    NOT NULL,
			article_id       INTEGER NOT NULL,
			author_id        INTEGER NOT NULL,
			author_name      TEXT    NOT NULL,
			publication_time INTEGER NOT NULL,
			tags             TEXT    NOT NULL,
			source_url       TEXT    NOT NULL,

			content          TEXT    NULL,
			read_count       INTEGER NULL,
			comment_count    INTEGER NULL,
	
			PRIMARY KEY (article_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sohu_news_comments (
			article_id    INTEGER NOT NULL,
			comment_id    INTEGER PRIMARY KEY,
			creation_time INTEGER NOT NULL,
			reference_ids INTEGER NULL,
			content       TEXT    NOT NULL,
			user_id       TEXT    NOT NULL,
	
			FOREIGN KEY (article_id) REFERENCES sohu_news  (article_id),
			FOREIGN KEY (user_id)    REFERENCES sohu_users (user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sohu_users (
			user_id  TEXT NOT NULL,
			nickname TEXT NOT NULL,
	
			PRIMARY KEY (user_id)
		)`,
	} {
		if _, err = db.Exec(stmt); err != nil {
			panic(err)
		}
	}
	c.Global["db"] = db
	c.Global["db-lock"] = &sync.Mutex{}
	defer db.Close()
	if c.Global["location"], err = time.LoadLocation("Asia/Shanghai"); err != nil {
		panic(err)
	}

	if err = c.RegisterTaskTypes(FeedTask{}, ArticleTask{}, ArticleCommentsTask{}, PVTask{}); err != nil {
		panic(err)
	}

	if err = c.Request(FeedTask{
		SceneID: 1460,
		Page:    1,
	}); err != nil {
		panic(err)
	}

	c.Run()

	var errs = c.WaitQuit()
	if errs != nil {
		fmt.Printf("%#v", errs)
	}

}
