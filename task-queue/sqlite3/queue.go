package s3queue

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	// sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

// Queue 是一个使用 SQLite 实现的队列
type Queue struct {
	db     *sql.DB
	dbLock sync.Mutex

	taskTypeNameToIDMap map[string]int
	taskTypeIDToNameMap map[int]string
	taskTypeMapLock     sync.Mutex

	iFactory     int
	iFactoryLock sync.Mutex

	defaultMaxAttempts int
}

// DefaultDefaultMaxAttempts 队列是默认的默认尝试次数
const DefaultDefaultMaxAttempts = 3

// Open 打开一个新的队列
func Open(path string) (q *Queue, err error) {
	q = new(Queue)
	if q.db, err = sql.Open("sqlite3", path); err != nil {
		panic(err)
	}

	var count int
	if err = q.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'queue_meta'`).Scan(&count); err != nil {
		panic(err)
	}

	if count == 0 {
		for _, stmt := range []string{
			// 元数据表, 存放元数据
			`CREATE TABLE queue_meta (
				meta_key   NOT NULL,
				meta_value NOT NULL
			)`,
			fmt.Sprintf(`INSERT INTO queue_meta (meta_key, meta_value) VALUES ('version', '1'),
			('task_type_map', '{}'),
			('i_factory', 1),
			('default_maximum_attempts', %d)`, DefaultDefaultMaxAttempts),

			// 任务表, 存放所有任务
			`CREATE TABLE tasks (
				task_id      INTEGER NOT NULL,

				task_type_id INTEGER NOT NULL,
				task_data    BLOB    NOT NULL,

				task_status  INTEGER NOT NULL, -- 等待/冻结/处理中的任务在此处都标记为等待

				PRIMARY KEY (task_id),
				UNIQUE(task_type_id, task_data)
			)`,

			// 队列表, 存放处于队列中的任务
			`CREATE TABLE pending_queue (
				task_id       INTEGER NOT NULL, -- 可正可负, 且决定队列中的先后顺序
				-- TODO: item_index
				item_priority INTEGER NOT NULL, -- 当其为负值时, 代表已被冻结

				item_remaining_attempts INTEGER NOT NULL, -- 当其为负值时, 代表正在被处理

				PRIMARY KEY (task_id)
			)`,
			`CREATE UNIQUE INDEX index_priority_id ON pending_queue (item_priority DESC, task_id ASC)`,
		} {

			if _, err = q.db.Exec(stmt); err != nil {
				panic(err)
			}
		}

	} else {
		// 将可能出现的因为意外中断而导致状态异常的任务的状态恢复正常
		if _, err = q.db.Exec(`UPDATE pending_queue SET item_remaining_attempts = -item_remaining_attempts WHERE item_remaining_attempts < 0`); err != nil {
			panic(err)
		}
	}

	{
		var taskTypeMapJSON []byte
		if err = q.db.QueryRow(`SELECT meta_value FROM queue_meta WHERE meta_key = 'task_type_map'`).Scan(&taskTypeMapJSON); err != nil {
			panic(err)
		}
		if err = json.Unmarshal(taskTypeMapJSON, &q.taskTypeNameToIDMap); err != nil {
			panic(err)
		} else {
			q.taskTypeIDToNameMap = make(map[int]string)
			for name, id := range q.taskTypeNameToIDMap {
				q.taskTypeIDToNameMap[id] = name
			}
		}
	}
	if err = q.db.QueryRow(`SELECT meta_value FROM queue_meta WHERE meta_key = 'i_factory'`).Scan(&q.iFactory); err != nil {
		panic(err)
	}

	if err = q.db.QueryRow(`SELECT meta_value FROM queue_meta WHERE meta_key = 'default_maximum_attempts'`).Scan(&q.defaultMaxAttempts); err != nil {
		panic(err)
	}

	return q, nil
}

// OpenInMemory 将队列开在内存
func OpenInMemory() (q *Queue, err error) {
	return Open(":memory:")
}

// Close 关闭队列
func (q *Queue) Close() (err error) {
	if err = q.updateDBTaskTypeMeta(false); err != nil {
		return err
	}

	if _, err = q.db.Exec(`UPDATE queue_meta SET meta_value = ? WHERE meta_key = 'i_factory'`, q.iFactory); err != nil {
		return err
	}

	if _, err = q.db.Exec(`UPDATE queue_meta SET meta_value = ? WHERE meta_key = 'default_maximum_attempts'`, q.defaultMaxAttempts); err != nil {
		return err
	}

	return q.db.Close()
}

func (q *Queue) nextI() (i int) {
	q.iFactoryLock.Lock()
	i = q.iFactory
	q.iFactory++
	q.iFactoryLock.Unlock()
	return i
}

// SetDefaultMaxAttempts 设置项目放入队列时, 默认的最大尝试次数
func (q *Queue) SetDefaultMaxAttempts(i int) {
	q.defaultMaxAttempts = i
}

// GetDefaultMaxAttempts 获取默认的最大尝试次数
func (q *Queue) GetDefaultMaxAttempts() (i int) {
	return q.defaultMaxAttempts
}

func (q *Queue) getTypeID(typeName string, alreadyLocked bool) (typeID int) {
	q.taskTypeMapLock.Lock()
	defer q.taskTypeMapLock.Unlock()

	if _typeID, exists := q.taskTypeNameToIDMap[typeName]; !exists {
		typeID = q.nextI()
		q.taskTypeNameToIDMap[typeName] = typeID
		q.taskTypeIDToNameMap[typeID] = typeName
		if err := q.updateDBTaskTypeMeta(alreadyLocked); err != nil {
			panic(err)
		}
	} else {
		typeID = _typeID
	}
	return typeID
}

func (q *Queue) getTypeName(typeID int) (typeName string) {
	return q.taskTypeIDToNameMap[typeID]
}

func (q *Queue) updateDBTaskTypeMeta(alreadyLocked bool) (err error) {
	var taskTypeMapJSON []byte
	if taskTypeMapJSON, err = json.Marshal(q.taskTypeNameToIDMap); err != nil {
		return err
	}
	if !alreadyLocked {
		q.dbLock.Lock()
	}
	_, err = q.db.Exec(`UPDATE queue_meta SET meta_value = ? WHERE meta_key = 'task_type_map'`, taskTypeMapJSON)
	if !alreadyLocked {
		q.dbLock.Unlock()
	}
	return err
}
