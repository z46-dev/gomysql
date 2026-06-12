package gosqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestBeginDefaultsToSingleConnectionPool(t *testing.T) {
	driver, err := Begin(filepath.Join(t.TempDir(), "default.sqlite"))
	if err != nil {
		t.Fatalf("failed to begin database: %v", err)
	}
	defer func() {
		if err := driver.Close(); err != nil {
			t.Fatalf("failed to close database: %v", err)
		}
	}()

	if got := driver.db.Stats().MaxOpenConnections; got != 1 {
		t.Fatalf("expected default max open connections to be 1, got %d", got)
	}
}

func TestBeginWithConnectionPoolEnablesForeignKeysOnEveryConnection(t *testing.T) {
	driver, err := Begin(filepath.Join(t.TempDir(), "pooled.sqlite"), WithConnectionPool(2, 2))
	if err != nil {
		t.Fatalf("failed to begin database: %v", err)
	}
	defer func() {
		if err := driver.Close(); err != nil {
			t.Fatalf("failed to close database: %v", err)
		}
	}()

	if got := driver.db.Stats().MaxOpenConnections; got != 2 {
		t.Fatalf("expected max open connections to be 2, got %d", got)
	}

	ctx := context.Background()

	conn1, err := driver.db.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to reserve first connection: %v", err)
	}
	defer conn1.Close()

	conn2, err := driver.db.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to reserve second connection: %v", err)
	}
	defer conn2.Close()

	for i, conn := range []*sql.Conn{conn1, conn2} {
		enabled, err := foreignKeysEnabled(ctx, conn)
		if err != nil {
			t.Fatalf("failed to inspect foreign_keys on connection %d: %v", i+1, err)
		}

		if !enabled {
			t.Fatalf("expected foreign_keys to be enabled on connection %d", i+1)
		}
	}
}

func foreignKeysEnabled(ctx context.Context, conn *sql.Conn) (bool, error) {
	var enabled int
	if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys;").Scan(&enabled); err != nil {
		return false, err
	}

	return enabled == 1, nil
}
