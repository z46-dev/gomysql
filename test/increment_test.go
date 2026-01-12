package test

import (
	"testing"

	"github.com/z46-dev/gomysql"
)

type PrimaryKeyIncrementDoc struct {
	ID   int    `gomysql:"id,primary,increment"`
	Name string `gomysql:"name,unique"`
}

type SecondaryKeyIncrementDoc struct {
	Name string `gomysql:"name,primary,unique"`
	Age  int    `gomysql:"age,unique,increment"`
}

func TestPKeyIncrement(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[PrimaryKeyIncrementDoc]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(PrimaryKeyIncrementDoc{}); err != nil {
		t.Fatalf("failed to register Document struct: %v", err)
	}

	var item = &PrimaryKeyIncrementDoc{
		Name: "Test Item",
	}

	if err = handler.Insert(item); err != nil {
		t.Fatalf("failed to insert item: %v", err)
	}

	var fetchedItem *PrimaryKeyIncrementDoc
	if fetchedItem, err = handler.Select(item.ID); err != nil {
		t.Fatalf("failed to select item: %v", err)
	}

	if fetchedItem.ID != 1 {
		t.Fatalf("expected ID to be 1, got %d", fetchedItem.ID)
	}

	if fetchedItem.Name != item.Name {
		t.Fatalf("expected Name to be %s, got %s", item.Name, fetchedItem.Name)
	}

	var item2 = &PrimaryKeyIncrementDoc{
		Name: "Test Item 2",
	}

	if err = handler.Insert(item2); err != nil {
		t.Fatalf("failed to insert item2: %v", err)
	}

	var fetchedItem2 *PrimaryKeyIncrementDoc
	if fetchedItem2, err = handler.Select(item2.ID); err != nil {
		t.Fatalf("failed to select item2: %v", err)
	}

	if fetchedItem2.ID != 2 {
		t.Fatalf("expected ID to be 2, got %d", fetchedItem2.ID)
	}

	if fetchedItem2.Name != item2.Name {
		t.Fatalf("expected Name to be %s, got %s", item2.Name, fetchedItem2.Name)
	}
}

func TestSKeyIncrement(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[SecondaryKeyIncrementDoc]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(SecondaryKeyIncrementDoc{}); err != nil {
		t.Fatalf("failed to register Document struct: %v", err)
	}

	var item = &SecondaryKeyIncrementDoc{
		Name: "Test Item",
	}

	if err = handler.Insert(item); err != nil {
		t.Fatalf("failed to insert item: %v", err)
	}

	var fetchedItem *SecondaryKeyIncrementDoc
	if fetchedItem, err = handler.Select(item.Name); err != nil {
		t.Fatalf("failed to select item: %v", err)
	}

	if fetchedItem.Age != 1 {
		t.Fatalf("expected Age to be 1, got %d", fetchedItem.Age)
	}

	if fetchedItem.Name != item.Name {
		t.Fatalf("expected Name to be %s, got %s", item.Name, fetchedItem.Name)
	}

	var item2 = &SecondaryKeyIncrementDoc{
		Name: "Test Item 2",
	}

	if err = handler.Insert(item2); err != nil {
		t.Fatalf("failed to insert item2: %v", err)
	}

	var fetchedItem2 *SecondaryKeyIncrementDoc
	if fetchedItem2, err = handler.Select(item2.Name); err != nil {
		t.Fatalf("failed to select item2: %v", err)
	}

	if fetchedItem2.Age != 2 {
		t.Fatalf("expected Age to be 2, got %d", fetchedItem2.Age)
	}

	if fetchedItem2.Name != item2.Name {
		t.Fatalf("expected Name to be %s, got %s", item2.Name, fetchedItem2.Name)
	}
}
