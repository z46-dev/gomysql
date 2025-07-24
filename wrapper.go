package gomysql

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

var DB *Driver

type Driver struct {
	db       *sql.DB
	lock     *sync.RWMutex
	filePath string
}

func Begin(dbPath string) (err error) {
	if DB != nil {
		err = ErrDatabaseInitialized
		return
	}

	var db *sql.DB

	if db, err = sql.Open("sqlite", dbPath); err != nil {
		return
	} else {
		DB = &Driver{
			db:       db,
			lock:     &sync.RWMutex{},
			filePath: dbPath,
		}
	}

	return
}

func Close() (err error) {
	if DB == nil {
		err = ErrDatabaseNotInitialized
		return
	}

	DB.lock.Lock()
	defer DB.lock.Unlock()
	if err = DB.db.Close(); err != nil {
		return
	}

	DB = nil
	return
}
