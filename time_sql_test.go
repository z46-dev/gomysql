package gomysql

import (
	"strings"
	"testing"
	"time"
)

type TimeColumnItem struct {
	ID        int       `gomysql:"id,primary"`
	Timestamp time.Time `gomysql:"timestamp"`
}

type LegacyTimeItem struct {
	ID        int       `gomysql:"id,primary"`
	CreatedAt time.Time `gomysql:"created_at"`
}

func withRootTestDB(t *testing.T, fn func()) {
	t.Helper()

	if err := Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	fn()
}

func TestTimeFieldUsesDateTimeColumn(t *testing.T) {
	withRootTestDB(t, func() {
		handler, err := Register(TimeColumnItem{})
		if err != nil {
			t.Fatalf("failed to register struct: %v", err)
		}

		cols, err := handler.db.tableColumns(handler.Name)
		if err != nil {
			t.Fatalf("failed to inspect columns: %v", err)
		}

		for _, col := range cols {
			if strings.EqualFold(col.Name, "timestamp") {
				if normalizeSQLType(col.Type) != "DATETIME" {
					t.Fatalf("expected timestamp column type DATETIME, got %s", col.Type)
				}
				return
			}
		}

		t.Fatalf("timestamp column not found")
	})
}

func TestMigrateLegacyTimeBlobColumn(t *testing.T) {
	withRootTestDB(t, func() {
		if _, err := DB.db.Exec("CREATE TABLE LegacyTimeItem (id INTEGER PRIMARY KEY, created_at BLOB);"); err != nil {
			t.Fatalf("failed to create legacy table: %v", err)
		}

		legacyTime := time.Date(2026, 2, 3, 4, 5, 6, 789123456, time.FixedZone("EST", -5*60*60))
		encoded, err := encodeTimeValue(legacyTime)
		if err != nil {
			t.Fatalf("failed to encode legacy time: %v", err)
		}

		if _, err := DB.db.Exec("INSERT INTO LegacyTimeItem (id, created_at) VALUES (?, ?);", 1, encoded); err != nil {
			t.Fatalf("failed to insert legacy row: %v", err)
		}

		handler, err := Register(LegacyTimeItem{})
		if err != nil {
			t.Fatalf("failed to register legacy struct: %v", err)
		}

		report, err := handler.Migrate(MigrationOptions{AllowDestructive: true})
		if err != nil {
			t.Fatalf("failed to migrate legacy time column: %v", err)
		}

		if !report.Rebuilt {
			t.Fatalf("expected migration rebuild for legacy time column")
		}

		row, err := handler.Select(1)
		if err != nil {
			t.Fatalf("failed to select migrated row: %v", err)
		}

		expected := legacyTime.UTC()
		if !row.CreatedAt.Equal(expected) {
			t.Fatalf("expected migrated time %s, got %s", expected, row.CreatedAt)
		}

		results, err := handler.SelectAllWithFilter(
			NewFilter().KeyCmp(handler.FieldByGoName("CreatedAt"), OpGreaterThan, expected.Add(-time.Second)),
		)
		if err != nil {
			t.Fatalf("failed to filter migrated time row: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected one migrated row from time comparison, got %d", len(results))
		}
	})
}
