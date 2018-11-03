# NYN

NYN 是一个使用 Go 语言编写的爬虫框架. 

## 使用例

参见[这里](./examples)

## TODO

### task-queue

* [ ] 使用 bolt(bbolt)/leveldb/badger 重写任务队列, 替换目前使用 SQLite 实现的任务队列

### data-packer

* objpack
  * [ ] 验证类型签名时, 忽略没有字段的内嵌结构体
  * [ ] 忽略/记录未导出的字段?

### task-manager

* [ ] 实现新旧类型的转换方法

### scheduler/simple-scheduler

* [ ] 实现支持 `ModeMonitor`

### nyn

* [ ] `crawler.GetShared()`, `crawler.SetShared()`, 同类型任务间共享的内容
* [ ] 用 TaskHandler 替代任务委托中传入的 `Crawler`
  * [ ] `Crawler.TaskLog` -> `TaskHandler.TaskLog`
* [ ] 考虑使用 swift 重写 (这个爬虫框架的思路在 Go 上显得很繁琐, 但配合 Swift 的 protocol + extension (+ 错误处理机制 )应该会清晰不少)

