package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/z46-dev/gosqlite"
)

func BenchmarkInsertSelect(b *testing.B) {
	withTestDB(b, func(driver *gosqlite.Driver) {
		handler, err := gosqlite.Register(driver, Document{})
		if err != nil {
			b.Fatalf("failed to register Document struct: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			doc := &Document{
				Title:        fmt.Sprintf("Bench Doc %d", i),
				Body:         "Benchmark body",
				Tags:         []string{"bench"},
				Creation:     time.Now(),
				BooleanField: i%2 == 0,
			}

			if err := handler.Insert(doc); err != nil {
				b.Fatalf("failed to insert document: %v", err)
			}

			if _, err := handler.Select(doc.ID); err != nil {
				b.Fatalf("failed to select document: %v", err)
			}
		}

		b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
	})
}

func BenchmarkSelectAllWithFilter(b *testing.B) {
	withTestDB(b, func(driver *gosqlite.Driver) {
		handler, err := gosqlite.Register(driver, Document{})
		if err != nil {
			b.Fatalf("failed to register Document struct: %v", err)
		}

		for i := 0; i < 1000; i++ {
			doc := &Document{
				Title:    fmt.Sprintf("Bench Doc %d", i),
				Body:     "Benchmark body",
				Tags:     []string{"bench"},
				Creation: time.Now().Add(time.Duration(i) * time.Second),
			}
			if err := handler.Insert(doc); err != nil {
				b.Fatalf("failed to insert document: %v", err)
			}
		}

		filter := gosqlite.NewFilter().
			KeyCmp(handler.FieldByGoName("Title"), gosqlite.OpLike, "%Bench%").
			And().
			KeyCmp(handler.FieldByGoName("ID"), gosqlite.OpGreaterThan, 100)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := handler.SelectAllWithFilter(filter); err != nil {
				b.Fatalf("failed to select with filter: %v", err)
			}
		}

		b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
	})
}

func BenchmarkUpdateWithFilterReturning(b *testing.B) {
	withTestDB(b, func(driver *gosqlite.Driver) {
		handler, err := gosqlite.Register(driver, Account{})
		if err != nil {
			b.Fatalf("failed to register Account struct: %v", err)
		}

		for i := 0; i < 500; i++ {
			account := &Account{
				Username: fmt.Sprintf("user-%d", i),
				Money:    1000,
			}
			if err := handler.Insert(account); err != nil {
				b.Fatalf("failed to insert account: %v", err)
			}
		}

		filter := gosqlite.NewFilter().
			KeyCmp(handler.FieldByGoName("Money"), gosqlite.OpGreaterThanOrEqual, 1000)
		returning := []*gosqlite.RegisteredStructField{
			handler.FieldByGoName("Money"),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := handler.UpdateWithFilterReturning(
				filter,
				returning,
				gosqlite.SetSub(handler.FieldByGoName("Money"), 1),
			); err != nil {
				b.Fatalf("failed to update with returning: %v", err)
			}
		}

		b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
	})
}
