package test

import (
	"math/rand/v2"
	"slices"
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
	if results, err = handler.SelectAllWithFilter(gomysql.NewFilter().KeyCmp(handler.FieldByGoName("Title"), gomysql.OpLike, "%even%")); err != nil {
		t.Fatalf("failed to select all documents with filter: %v", err)
	}

	assert.Len(t, results, 64, "expected 64 documents with 'even' tag")

	for _, doc := range results {
		assert.Contains(t, doc.Title, "even", "document should contain 'even' tag")
	}
}

func TestOrderingFilter(t *testing.T) {
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
		docs[i] = Document{
			Title:    "Filter Test Document",
			Body:     "This document is used for testing filters.",
			Creation: time.Now().Add(time.Duration(i) * time.Hour),
		}
	}

	// Randomize the docs
	slices.SortFunc(docs, func(a, b Document) int {
		if rand.Float64() < 0.5 {
			return -1 // a comes before b
		}
		return 1 // b comes before a
	})

	for i := range docs {
		if err = handler.Insert(&docs[i]); err != nil {
			t.Fatalf("failed to insert document %d: %v", i, err)
		}
	}

	// Get in order
	var results []*Document
	if results, err = handler.SelectAllWithFilter(gomysql.NewFilter().Ordering(handler.FieldByGoName("Creation"), true)); err != nil {
		t.Fatalf("failed to select all documents with ordering: %v", err)
	}

	assert.Len(t, results, 128, "expected 128 documents")
	for i := 1; i < len(results); i++ {
		assert.LessOrEqual(t, results[i-1].Creation.Unix(), results[i].Creation.Unix(), "documents should be ordered by creation time")
	}
}
