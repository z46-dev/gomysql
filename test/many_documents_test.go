package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gomysql"
)

func TestManyDocs(t *testing.T) {
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

	var docs []Document = make([]Document, 1024)
	for i := range docs {
		docs[i] = Document{
			Title:    fmt.Sprintf("Doc #%4X", i),
			Body:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			Tags:     []string{"tag1", "tag2", "tag3"},
			Creation: time.Now(),
		}

		if err = handler.Insert(&docs[i]); err != nil {
			t.Fatalf("failed to insert document %d: %v", i, err)
		}
	}

	var primaryKeys []any
	if primaryKeys, err = handler.List(); err != nil {
		t.Fatalf("failed to list primary keys: %v", err)
	}

	assert.Len(t, primaryKeys, len(docs), "expected to retrieve all primary keys")
	for i, pk := range primaryKeys {
		assert.EqualValues(t, docs[i].ID, pk, "primary key mismatch at index %d", i)
	}

	var fetchedDocs []*Document
	if fetchedDocs, err = handler.SelectAll(); err != nil {
		t.Fatalf("failed to select all documents: %v", err)
	}

	assert.Len(t, fetchedDocs, len(docs), "expected to retrieve all documents")
	for i, doc := range fetchedDocs {
		twoDocsMatch(t, &docs[i], doc)
	}
}
