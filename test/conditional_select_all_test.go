package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gomysql"
)

func TestFilters(t *testing.T) {
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

	var docs []Document = make([]Document, 128)
	for i := range docs {
		var filterTag string
		if i%2 == 0 {
			filterTag = "even"
		} else {
			filterTag = "odd"
		}

		docs[i] = Document{
			Title:    "Filter Test Document " + filterTag,
			Body:     "This document is used for testing filters.",
			Tags:     []string{"filter", "test"},
			Creation: time.Now(),
		}

		if err = handler.Insert(&docs[i]); err != nil {
			t.Fatalf("failed to insert document %d: %v", i, err)
		}
	}

	var results []*Document
	if results, err = handler.SelectAllWithFilter(gomysql.NewFilter().KeyCmp(handler.FieldByGoName("Title"), gomysql.OpLike, "'%even%'")); err != nil {
		t.Fatalf("failed to select all documents with filter: %v", err)
	}

	assert.Len(t, results, 64, "expected 64 documents with 'even' tag")

	for _, doc := range results {
		assert.Contains(t, doc.Title, "even", "document should contain 'even' tag")
	}
}
