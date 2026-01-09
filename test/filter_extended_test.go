package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gomysql"
)

func TestFilterGroupingAndIn(t *testing.T) {
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

	titles := []string{
		"Grouped even doc",
		"Grouped odd doc",
		"Grouped other doc",
		"Grouped even doc 2",
	}

	var ids []int
	for _, title := range titles {
		doc := &Document{
			Title:    title,
			Body:     "Grouped filter test",
			Tags:     []string{"group", "filter"},
			Creation: time.Now(),
		}

		if err = handler.Insert(doc); err != nil {
			t.Fatalf("failed to insert document: %v", err)
		}
		ids = append(ids, doc.ID)
	}

	filter := gomysql.NewFilter().
		OpenGroup().
		KeyCmp(handler.FieldByGoName("Title"), gomysql.OpLike, "%even%").
		Or().
		KeyCmp(handler.FieldByGoName("Title"), gomysql.OpLike, "%odd%").
		CloseGroup().
		And().
		KeyCmp(handler.FieldByGoName("ID"), gomysql.OpIn, []int{ids[0], ids[1], ids[3]})

	results, err := handler.SelectAllWithFilter(filter)
	if err != nil {
		t.Fatalf("failed to select with grouped filter: %v", err)
	}

	assert.Len(t, results, 3, "expected three documents from grouped filter")
	for _, doc := range results {
		assert.Contains(t, doc.Title, "Grouped", "document title should match base")
	}
}

func TestFilterNotIn(t *testing.T) {
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

	var ids []int
	for i := 0; i < 4; i++ {
		doc := &Document{
			Title:    "Not in filter doc",
			Body:     "Not in filter test",
			Tags:     []string{"notin"},
			Creation: time.Now(),
		}

		if err = handler.Insert(doc); err != nil {
			t.Fatalf("failed to insert document %d: %v", i, err)
		}
		ids = append(ids, doc.ID)
	}

	filter := gomysql.NewFilter().
		KeyCmp(handler.FieldByGoName("ID"), gomysql.OpNotIn, []int{ids[1], ids[2]})

	results, err := handler.SelectAllWithFilter(filter)
	if err != nil {
		t.Fatalf("failed to select with not-in filter: %v", err)
	}

	assert.Len(t, results, 2, "expected two documents after not-in filter")
}
