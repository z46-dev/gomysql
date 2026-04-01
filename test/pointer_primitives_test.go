package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gomysql"
)

type PointerPrimitiveDoc struct {
	ID       int      `gomysql:"id,primary,increment"`
	Age      *int     `gomysql:"age"`
	Score    *float64 `gomysql:"score"`
	Enabled  *bool    `gomysql:"enabled"`
	Nickname *string  `gomysql:"nickname"`
}

func intPtr(v int) *int             { return &v }
func float64Ptr(v float64) *float64 { return &v }
func boolPtr(v bool) *bool          { return &v }
func stringPtr(v string) *string    { return &v }

func TestPointerPrimitiveInsertAndNullFilter(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[PointerPrimitiveDoc]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(PointerPrimitiveDoc{}); err != nil {
		t.Fatalf("failed to register PointerPrimitiveDoc struct: %v", err)
	}

	withValues := &PointerPrimitiveDoc{
		Age:      intPtr(42),
		Score:    float64Ptr(9.5),
		Enabled:  boolPtr(true),
		Nickname: stringPtr("neo"),
	}
	withNulls := &PointerPrimitiveDoc{}

	if err = handler.Insert(withValues); err != nil {
		t.Fatalf("failed to insert document with values: %v", err)
	}
	if err = handler.Insert(withNulls); err != nil {
		t.Fatalf("failed to insert document with nulls: %v", err)
	}

	fetchedValues, err := handler.Select(withValues.ID)
	if err != nil {
		t.Fatalf("failed to select document with values: %v", err)
	}
	assert.NotNil(t, fetchedValues.Age)
	assert.Equal(t, 42, *fetchedValues.Age)
	assert.NotNil(t, fetchedValues.Score)
	assert.InDelta(t, 9.5, *fetchedValues.Score, 0.001)
	assert.NotNil(t, fetchedValues.Enabled)
	assert.True(t, *fetchedValues.Enabled)
	assert.NotNil(t, fetchedValues.Nickname)
	assert.Equal(t, "neo", *fetchedValues.Nickname)

	fetchedNulls, err := handler.Select(withNulls.ID)
	if err != nil {
		t.Fatalf("failed to select document with nulls: %v", err)
	}
	assert.Nil(t, fetchedNulls.Age)
	assert.Nil(t, fetchedNulls.Score)
	assert.Nil(t, fetchedNulls.Enabled)
	assert.Nil(t, fetchedNulls.Nickname)

	nullResults, err := handler.SelectAllWithFilter(
		gomysql.NewFilter().
			KeyCmp(handler.FieldByGoName("Age"), gomysql.OpIsNull, nil),
	)
	if err != nil {
		t.Fatalf("failed to select null age rows: %v", err)
	}
	if assert.Len(t, nullResults, 1) {
		assert.Equal(t, withNulls.ID, nullResults[0].ID)
		assert.Nil(t, nullResults[0].Age)
	}

	notNullResults, err := handler.SelectAllWithFilter(
		gomysql.NewFilter().
			KeyCmp(handler.FieldByGoName("Age"), gomysql.OpIsNotNull, nil),
	)
	if err != nil {
		t.Fatalf("failed to select non-null age rows: %v", err)
	}
	if assert.Len(t, notNullResults, 1) {
		assert.Equal(t, withValues.ID, notNullResults[0].ID)
		assert.NotNil(t, notNullResults[0].Age)
		assert.Equal(t, 42, *notNullResults[0].Age)
	}
}

func TestPointerPrimitiveUpdateWithFilter(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[PointerPrimitiveDoc]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(PointerPrimitiveDoc{}); err != nil {
		t.Fatalf("failed to register PointerPrimitiveDoc struct: %v", err)
	}

	item := &PointerPrimitiveDoc{
		Age:      intPtr(21),
		Score:    float64Ptr(1.25),
		Enabled:  boolPtr(false),
		Nickname: stringPtr("start"),
	}

	if err = handler.Insert(item); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	rows, err := handler.UpdateWithFilter(
		gomysql.NewFilter().KeyCmp(handler.FieldByGoName("ID"), gomysql.OpEqual, item.ID),
		gomysql.SetField(handler.FieldByGoName("Age"), nil),
		gomysql.SetField(handler.FieldByGoName("Nickname"), "done"),
	)
	if err != nil {
		t.Fatalf("failed to update document: %v", err)
	}
	assert.EqualValues(t, 1, rows)

	fetched, err := handler.Select(item.ID)
	if err != nil {
		t.Fatalf("failed to select updated document: %v", err)
	}
	assert.Nil(t, fetched.Age)
	if assert.NotNil(t, fetched.Nickname) {
		assert.Equal(t, "done", *fetched.Nickname)
	}

	returned, err := handler.UpdateWithFilterReturning(
		gomysql.NewFilter().KeyCmp(handler.FieldByGoName("ID"), gomysql.OpEqual, item.ID),
		[]*gomysql.RegisteredStructField{
			handler.FieldByGoName("Age"),
			handler.FieldByGoName("Nickname"),
		},
		gomysql.SetField(handler.FieldByGoName("Nickname"), nil),
	)
	if err != nil {
		t.Fatalf("failed to update document with returning: %v", err)
	}
	if assert.Len(t, returned, 1) {
		assert.Nil(t, returned[0]["age"])
		assert.Nil(t, returned[0]["nickname"])
	}
}
