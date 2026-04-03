package test

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gomysql"
)

func TestCountWithFilter(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[Document]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(Document{}); err != nil {
		t.Fatalf("failed to register Document struct: %v", err)
	}

	base := time.Now().UTC().Truncate(time.Second)
	for i := 0; i < 6; i++ {
		doc := &Document{
			Title:        "Count me",
			Body:         "count test",
			Tags:         []string{"count"},
			Creation:     base.Add(time.Duration(i) * time.Minute),
			BooleanField: i%2 == 0,
		}

		if err = handler.Insert(doc); err != nil {
			t.Fatalf("failed to insert document %d: %v", i, err)
		}
	}

	total, err := handler.Count()
	if err != nil {
		t.Fatalf("failed to count all rows: %v", err)
	}
	assert.EqualValues(t, 6, total, "expected total row count")

	filtered, err := handler.CountWithFilter(
		gomysql.NewFilter().
			KeyCmp(handler.FieldByGoName("BooleanField"), gomysql.OpEqual, true),
	)
	if err != nil {
		t.Fatalf("failed to count filtered rows: %v", err)
	}
	assert.EqualValues(t, 3, filtered, "expected filtered row count")

	limited, err := handler.CountWithFilter(
		gomysql.NewFilter().
			Ordering(handler.FieldByGoName("Creation"), true).
			Limit(2),
	)
	if err != nil {
		t.Fatalf("failed to count limited rows: %v", err)
	}
	assert.EqualValues(t, 2, limited, "expected count to respect limit")
}

func TestDeleteWithFilterOldestRows(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[Document]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(Document{}); err != nil {
		t.Fatalf("failed to register Document struct: %v", err)
	}

	base := time.Now().UTC().Truncate(time.Second)
	docs := []Document{
		{Title: "third oldest", Body: "delete filter test", Tags: []string{"delete"}, Creation: base.Add(2 * time.Hour)},
		{Title: "oldest", Body: "delete filter test", Tags: []string{"delete"}, Creation: base},
		{Title: "newest", Body: "delete filter test", Tags: []string{"delete"}, Creation: base.Add(4 * time.Hour)},
		{Title: "second oldest", Body: "delete filter test", Tags: []string{"delete"}, Creation: base.Add(1 * time.Hour)},
		{Title: "second newest", Body: "delete filter test", Tags: []string{"delete"}, Creation: base.Add(3 * time.Hour)},
	}

	for i := range docs {
		if err = handler.Insert(&docs[i]); err != nil {
			t.Fatalf("failed to insert document %d: %v", i, err)
		}
	}

	deleted, err := handler.DeleteWithFilter(
		gomysql.NewFilter().
			Ordering(handler.FieldByGoName("Creation"), true).
			Limit(2),
	)
	if err != nil {
		t.Fatalf("failed to delete oldest rows: %v", err)
	}
	assert.EqualValues(t, 2, deleted, "expected two deleted rows")

	remainingCount, err := handler.Count()
	if err != nil {
		t.Fatalf("failed to count remaining rows: %v", err)
	}
	assert.EqualValues(t, 3, remainingCount, "expected three remaining rows")

	remaining, err := handler.SelectAllWithFilter(
		gomysql.NewFilter().
			Ordering(handler.FieldByGoName("Creation"), true),
	)
	if err != nil {
		t.Fatalf("failed to select remaining rows: %v", err)
	}

	assert.Len(t, remaining, 3, "expected three remaining documents")

	var titles []string
	for _, doc := range remaining {
		titles = append(titles, doc.Title)
	}

	slices.Sort(titles)
	assert.Equal(t, []string{"newest", "second newest", "third oldest"}, titles, "expected oldest rows to be deleted")
}
