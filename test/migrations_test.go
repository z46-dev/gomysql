package test

import (
	"testing"

	"github.com/z46-dev/gomysql"
	v1 "github.com/z46-dev/gomysql/test/migrationv1"
	v2 "github.com/z46-dev/gomysql/test/migrationv2"
)

func withTestDB(t *testing.T, fn func()) {
	t.Helper()

	if err := gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	fn()
}

func TestMigrationAddColumn(t *testing.T) {
	withTestDB(t, func() {
		v1Handler, err := gomysql.Register(v1.AddItem{})
		if err != nil {
			t.Fatalf("failed to register v1 struct: %v", err)
		}

		item := &v1.AddItem{Name: "alpha"}
		if err := v1Handler.Insert(item); err != nil {
			t.Fatalf("failed to insert v1 item: %v", err)
		}

		v2Handler, err := gomysql.Register(v2.AddItem{})
		if err != nil {
			t.Fatalf("failed to register v2 struct: %v", err)
		}

		if _, err := v2Handler.Migrate(gomysql.MigrationOptions{}); err != nil {
			t.Fatalf("failed to migrate add column: %v", err)
		}

		got, err := v2Handler.Select(item.ID)
		if err != nil {
			t.Fatalf("failed to select after migration: %v", err)
		}

		if got.Age != 0 {
			t.Fatalf("expected default Age to be 0, got %d", got.Age)
		}
	})
}

func TestMigrationDropColumn(t *testing.T) {
	withTestDB(t, func() {
		v1Handler, err := gomysql.Register(v1.DropItem{})
		if err != nil {
			t.Fatalf("failed to register v1 struct: %v", err)
		}

		item := &v1.DropItem{Name: "bravo", Age: 22}
		if err := v1Handler.Insert(item); err != nil {
			t.Fatalf("failed to insert v1 item: %v", err)
		}

		v2Handler, err := gomysql.Register(v2.DropItem{})
		if err != nil {
			t.Fatalf("failed to register v2 struct: %v", err)
		}

		report, err := v2Handler.Migrate(gomysql.MigrationOptions{AllowDestructive: true})
		if err != nil {
			t.Fatalf("failed to migrate drop column: %v", err)
		}

		if !report.Rebuilt {
			t.Fatalf("expected rebuild for drop migration")
		}

		if len(report.DroppedColumns) != 1 || report.DroppedColumns[0] != "age" {
			t.Fatalf("expected dropped column to be age, got %v", report.DroppedColumns)
		}

		if _, err := v2Handler.Select(item.ID); err != nil {
			t.Fatalf("failed to select after migration: %v", err)
		}

		if _, err := v1Handler.Select(item.ID); err == nil {
			t.Fatalf("expected old handler to fail after column drop")
		}
	})
}

func TestMigrationChangeType(t *testing.T) {
	withTestDB(t, func() {
		v1Handler, err := gomysql.Register(v1.TypeItem{})
		if err != nil {
			t.Fatalf("failed to register v1 struct: %v", err)
		}

		item := &v1.TypeItem{Active: 1}
		if err := v1Handler.Insert(item); err != nil {
			t.Fatalf("failed to insert v1 item: %v", err)
		}

		v2Handler, err := gomysql.Register(v2.TypeItem{})
		if err != nil {
			t.Fatalf("failed to register v2 struct: %v", err)
		}

		report, err := v2Handler.Migrate(gomysql.MigrationOptions{AllowDestructive: true})
		if err != nil {
			t.Fatalf("failed to migrate type change: %v", err)
		}

		if !report.Rebuilt {
			t.Fatalf("expected rebuild for type change")
		}

		if len(report.ChangedColumns) != 1 || report.ChangedColumns[0] != "active" {
			t.Fatalf("expected changed column to be active, got %v", report.ChangedColumns)
		}

		got, err := v2Handler.Select(item.ID)
		if err != nil {
			t.Fatalf("failed to select after migration: %v", err)
		}

		if !got.Active {
			t.Fatalf("expected Active to be true after migration")
		}
	})
}
