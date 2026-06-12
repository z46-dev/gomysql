package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gosqlite"
)

type Document struct {
	ID           int       `gosqlite:"id,primary,increment"`
	Title        string    `gosqlite:"title"`
	Body         string    `gosqlite:"body"`
	Tags         []string  `gosqlite:"tags"`
	Creation     time.Time `gosqlite:"creation"`
	BooleanField bool      `gosqlite:"boolean_field"`
}

func twoDocsMatch(t *testing.T, doc1, doc2 *Document) {
	assert.Equal(t, doc1.ID, doc2.ID, "IDs should match")
	assert.Equal(t, doc1.Title, doc2.Title, "Titles should match")
	assert.Equal(t, doc1.Body, doc2.Body, "Bodies should match")
	assert.ElementsMatch(t, doc1.Tags, doc2.Tags, "Tags should match")
	assert.WithinDuration(t, doc1.Creation, doc2.Creation, time.Second, "Creation times should be within 1 second")
	assert.Equal(t, doc1.BooleanField, doc2.BooleanField, "Boolean fields should match")
}

func withTestDB(t testing.TB, fn func(*gosqlite.Driver)) {
	t.Helper()

	driver, err := gosqlite.Begin(":memory:")
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := driver.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	fn(driver)
}

type MultiLayerStructEmbeddedStruct struct {
	Name    string `gosqlite:"name"`
	Content string `gosqlite:"content"`
}

type MultiLayerStruct struct {
	Name          string                          `gosqlite:"name,primary,unique"`
	Layer1        MultiLayerStructEmbeddedStruct  `gosqlite:"layer1"`
	Layer2        MultiLayerStructEmbeddedStruct  `gosqlite:"layer2"`
	PointerLayer1 *MultiLayerStructEmbeddedStruct `gosqlite:"pointer_layer1"`
	PointerLayer2 *MultiLayerStructEmbeddedStruct `gosqlite:"pointer_layer2"`
}

func twoMultiLayerStructsMatch(t *testing.T, doc1, doc2 *MultiLayerStruct) {
	assert.Equal(t, doc1.Name, doc2.Name, "Names should match")
	assert.Equal(t, doc1.Layer1.Name, doc2.Layer1.Name, "Layer 1 Names should match")
	assert.Equal(t, doc1.Layer1.Content, doc2.Layer1.Content, "Layer 1 Contents should match")
	assert.Equal(t, doc1.Layer2.Name, doc2.Layer2.Name, "Layer 2 Names should match")
	assert.Equal(t, doc1.Layer2.Content, doc2.Layer2.Content, "Layer 2 Contents should match")
	if doc1.PointerLayer1 != nil && doc2.PointerLayer1 != nil {
		assert.Equal(t, doc1.PointerLayer1.Name, doc2.PointerLayer1.Name, "Pointer Layer 1 Names should match")
		assert.Equal(t, doc1.PointerLayer1.Content, doc2.PointerLayer1.Content, "Pointer Layer 1 Contents should match")
	} else {
		assert.Nil(t, doc2.PointerLayer1)
	}
	if doc1.PointerLayer2 != nil && doc2.PointerLayer2 != nil {
		assert.Equal(t, doc1.PointerLayer2.Name, doc2.PointerLayer2.Name, "Pointer Layer 2 Names should match")
		assert.Equal(t, doc1.PointerLayer2.Content, doc2.PointerLayer2.Content, "Pointer Layer 2 Contents should match")
	} else {
		assert.Nil(t, doc2.PointerLayer2)
	}
}
