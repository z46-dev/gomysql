package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Document struct {
	ID       int       `gomysql:"id,primary,increment"`
	Title    string    `gomysql:"title"`
	Body     string    `gomysql:"body"`
	Tags     []string  `gomysql:"tags"`
	Creation time.Time `gomysql:"creation"`
}

func twoDocsMatch(t *testing.T, doc1, doc2 *Document) {
	assert.Equal(t, doc1.ID, doc2.ID, "IDs should match")
	assert.Equal(t, doc1.Title, doc2.Title, "Titles should match")
	assert.Equal(t, doc1.Body, doc2.Body, "Bodies should match")
	assert.ElementsMatch(t, doc1.Tags, doc2.Tags, "Tags should match")
	assert.WithinDuration(t, doc1.Creation, doc2.Creation, time.Second, "Creation times should be within 1 second")
}
