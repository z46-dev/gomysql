package gosqlite

import (
	"context"
	"database/sql"
	"net/url"
	"strings"
	"sync"

	sqlitedriver "modernc.org/sqlite"
)

type Driver struct {
	db       *sql.DB
	lock     *sync.RWMutex
	filePath string
}

type BeginOption func(*beginConfig)

type beginConfig struct {
	maxOpenConns int
	maxIdleConns int
}

const gosqliteForeignKeysHookParam = "_gosqlite_fk_hook"

var registerSQLiteHookOnce sync.Once

// WithConnectionPool overrides Begin's default single-connection sql.DB pool.
// Keep the default when using SQLite :memory: DSNs unless you intentionally use
// a shared in-memory URI, because separate connections can observe different databases.
func WithConnectionPool(maxOpenConns, maxIdleConns int) BeginOption {
	return func(cfg *beginConfig) {
		cfg.maxOpenConns = maxOpenConns
		cfg.maxIdleConns = maxIdleConns
	}
}

func Begin(dbPath string, opts ...BeginOption) (driver *Driver, err error) {
	var db *sql.DB
	cfg := beginConfig{
		maxOpenConns: 1,
		maxIdleConns: 1,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	registerSQLiteHook()

	if db, err = sql.Open("sqlite", withForeignKeysHook(dbPath)); err != nil {
		return
	} else {
		db.SetMaxOpenConns(cfg.maxOpenConns)
		db.SetMaxIdleConns(cfg.maxIdleConns)

		if err = db.Ping(); err != nil {
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

func (d *Driver) GetFilePath() (path string) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	path = d.filePath
	return
}

func registerSQLiteHook() {
	registerSQLiteHookOnce.Do(func() {
		sqlitedriver.RegisterConnectionHook(func(conn sqlitedriver.ExecQuerierContext, dsn string) error {
			if !hasForeignKeysHook(dsn) {
				return nil
			}

			_, err := conn.ExecContext(context.Background(), "PRAGMA foreign_keys = ON;", nil)
			return err
		})
	})
}

func withForeignKeysHook(dsn string) string {
	sep := "?"
	switch {
	case strings.Contains(dsn, "?") && !strings.HasSuffix(dsn, "?") && !strings.HasSuffix(dsn, "&"):
		sep = "&"
	case strings.HasSuffix(dsn, "?"), strings.HasSuffix(dsn, "&"):
		sep = ""
	}

	return dsn + sep + gosqliteForeignKeysHookParam + "=1"
}

func hasForeignKeysHook(dsn string) bool {
	queryIndex := strings.IndexByte(dsn, '?')
	if queryIndex == -1 || queryIndex == len(dsn)-1 {
		return false
	}

	values, err := url.ParseQuery(dsn[queryIndex+1:])
	if err != nil {
		return false
	}

	for _, value := range values[gosqliteForeignKeysHookParam] {
		if value == "1" {
			return true
		}
	}

	return false
}
