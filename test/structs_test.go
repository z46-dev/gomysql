package test

import (
	"testing"

	"github.com/z46-dev/gomysql"
)

func TestEmbeddedStructs(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[MultiLayerStruct]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(MultiLayerStruct{}); err != nil {
		t.Fatalf("failed to register Document struct: %v", err)
	}

	var item = &MultiLayerStruct{
		Name: "Test Item",
		Layer1: MultiLayerStructEmbeddedStruct{
			Name:    "Layer 1",
			Content: "Content for layer 1",
		},
		Layer2: MultiLayerStructEmbeddedStruct{
			Name:    "Layer 2",
			Content: "Content for layer 2",
		},
		PointerLayer1: nil,
		PointerLayer2: &MultiLayerStructEmbeddedStruct{
			Name:    "Pointer Layer 2",
			Content: "Content for pointer layer 2",
		},
	}

	if err = handler.Insert(item); err != nil {
		t.Fatalf("failed to insert item: %v", err)
	}

	var fetchedItem *MultiLayerStruct
	if fetchedItem, err = handler.Select(item.Name); err != nil {
		t.Fatalf("failed to select item: %v", err)
	}

	twoMultiLayerStructsMatch(t, item, fetchedItem)
}
