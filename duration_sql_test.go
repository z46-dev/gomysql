package gosqlite

import (
	"strings"
	"testing"
	"time"
)

type DurationColumnItem struct {
	ID      int           `gosqlite:"id,primary"`
	Timeout time.Duration `gosqlite:"timeout"`
}

func TestDurationFieldUsesIntegerColumn(t *testing.T) {
	withRootTestDB(t, func(driver *Driver) {
		handler, err := Register(driver, DurationColumnItem{})
		if err != nil {
			t.Fatalf("failed to register struct: %v", err)
		}

		cols, err := handler.driver.tableColumns(handler.Name)
		if err != nil {
			t.Fatalf("failed to inspect columns: %v", err)
		}

		for _, col := range cols {
			if strings.EqualFold(col.Name, "timeout") {
				if normalizeSQLType(col.Type) != "INTEGER" {
					t.Fatalf("expected timeout column type INTEGER, got %s", col.Type)
				}
				return
			}
		}

		t.Fatalf("timeout column not found")
	})
}

func TestDurationFieldRoundTripsAndFilters(t *testing.T) {
	withRootTestDB(t, func(driver *Driver) {
		handler, err := Register(driver, DurationColumnItem{})
		if err != nil {
			t.Fatalf("failed to register struct: %v", err)
		}

		expected := 90*time.Second + 250*time.Millisecond
		row := &DurationColumnItem{
			ID:      1,
			Timeout: expected,
		}

		if err := handler.Insert(row); err != nil {
			t.Fatalf("failed to insert duration row: %v", err)
		}

		selected, err := handler.Select(1)
		if err != nil {
			t.Fatalf("failed to select duration row: %v", err)
		}

		if selected.Timeout != expected {
			t.Fatalf("expected duration %s, got %s", expected, selected.Timeout)
		}

		results, err := handler.SelectAllWithFilter(
			NewFilter().
				KeyCmp(handler.FieldByGoName("Timeout"), OpGreaterThan, time.Minute).
				Ordering(handler.FieldByGoName("Timeout"), false),
		)
		if err != nil {
			t.Fatalf("failed to filter duration row: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected one duration row from comparison, got %d", len(results))
		}
	})
}
