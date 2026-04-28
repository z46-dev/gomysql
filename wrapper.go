package gomysql

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

type Driver struct {
	db       *sql.DB
	lock     *sync.RWMutex
	filePath string
}

func Begin(dbPath string) (driver *Driver, err error) {
	var db *sql.DB

	if db, err = sql.Open("sqlite", dbPath); err != nil {
		return
	} else {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)

		if _, err = db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
			_ = db.Close()
			return
		}

		driver = &Driver{
			db:       db,
			lock:     &sync.RWMutex{},
			filePath: dbPath,
		}
	}

	return
}

func (d *Driver) Close() (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if err = d.db.Close(); err != nil {
		return
	}

	d = nil
	return
}
