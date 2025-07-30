package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gomysql"
)

func TestDocStruct(t *testing.T) {
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

	var doc = &Document{
		Title:    "Test Document",
		Body:     "This is a test document.",
		Tags:     []string{"test", "document"},
		Creation: time.Now(),
	}

	if err = handler.Insert(doc); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	assert.Equal(t, 1, doc.ID, "ID should be 1 before insertion with autoincrement")

	var fetchedDoc *Document
	if fetchedDoc, err = handler.Select(doc.ID); err != nil {
		t.Fatalf("failed to select document: %v", err)
	}

	twoDocsMatch(t, doc, fetchedDoc)

	doc.Title = "Updated Document"
	if err = handler.Update(doc); err != nil {
		t.Fatalf("failed to update document: %v", err)
	}

	if fetchedDoc, err = handler.Select(doc.ID); err != nil {
		t.Fatalf("failed to select updated document: %v", err)
	}

	twoDocsMatch(t, doc, fetchedDoc)

	var list []any
	if list, err = handler.List(); err != nil {
		t.Fatalf("failed to list documents: %v", err)
	}

	assert.Equal(t, len(list), 1, "List should contain one document")
	assert.EqualValues(t, doc.ID, list[0], "List should contain the correct document ID")

	if err = handler.Delete(doc.ID); err != nil {
		t.Fatalf("failed to delete document: %v", err)
	}

	if fetchedDoc, err = handler.Select(doc.ID); err != nil {
		t.Fatalf("expected no error when selecting deleted document, got %v", err)
	} else if fetchedDoc != nil {
		t.Fatalf("expected fetched document to be nil after deletion, got %v", fetchedDoc)
	}

	assert.Nil(t, fetchedDoc, "Fetched document should be nil after deletion")
	if list, err = handler.List(); err != nil {
		t.Fatalf("failed to list documents after deletion: %v", err)
	}

	assert.Empty(t, list, "List should be empty after deletion")
}
