package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gosqlite"
)

type PermissionRecord struct {
	ID    int `gosqlite:"id,primary,increment"`
	Flags int `gosqlite:"flags"`
}

const (
	permissionRead = 1 << iota
	permissionWrite
	permissionAdmin
)

func TestFilterGroupingAndIn(t *testing.T) {
	withTestDB(t, func(driver *gosqlite.Driver) {
		var (
			err     error
			handler *gosqlite.RegisteredStruct[Document]
		)

		if handler, err = gosqlite.Register(driver, Document{}); err != nil {
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

		filter := gosqlite.NewFilter().
			OpenGroup().
			KeyCmp(handler.FieldByGoName("Title"), gosqlite.OpLike, "%even%").
			Or().
			KeyCmp(handler.FieldByGoName("Title"), gosqlite.OpLike, "%odd%").
			CloseGroup().
			And().
			KeyCmp(handler.FieldByGoName("ID"), gosqlite.OpIn, []int{ids[0], ids[1], ids[3]})

		results, err := handler.SelectAllWithFilter(filter)
		if err != nil {
			t.Fatalf("failed to select with grouped filter: %v", err)
		}

		assert.Len(t, results, 3, "expected three documents from grouped filter")
		for _, doc := range results {
			assert.Contains(t, doc.Title, "Grouped", "document title should match base")
		}
	})
}

func TestFilterBitwiseFlags(t *testing.T) {
	withTestDB(t, func(driver *gosqlite.Driver) {
		var (
			err     error
			handler *gosqlite.RegisteredStruct[PermissionRecord]
		)

		if handler, err = gosqlite.Register(driver, PermissionRecord{}); err != nil {
			t.Fatalf("failed to register PermissionRecord struct: %v", err)
		}

		records := []*PermissionRecord{
			{Flags: permissionRead},
			{Flags: permissionRead | permissionWrite},
			{Flags: permissionAdmin},
			{Flags: permissionRead | permissionWrite | permissionAdmin},
		}

		for _, record := range records {
			if err = handler.Insert(record); err != nil {
				t.Fatalf("failed to insert permission record: %v", err)
			}
		}

		flagsField := handler.FieldByGoName("Flags")
		readOrWrite := permissionRead | permissionWrite

		anyWrite, err := handler.SelectAllWithFilter(
			gosqlite.NewFilter().KeyHasAnyBits(flagsField, permissionWrite),
		)
		if err != nil {
			t.Fatalf("failed to select records with any write bit: %v", err)
		}
		assert.Len(t, anyWrite, 2, "expected records with the write bit set")

		allReadWrite, err := handler.SelectAllWithFilter(
			gosqlite.NewFilter().KeyHasAllBits(flagsField, readOrWrite),
		)
		if err != nil {
			t.Fatalf("failed to select records with all read/write bits: %v", err)
		}
		assert.Len(t, allReadWrite, 2, "expected records with read and write bits set")

		noAdmin, err := handler.SelectAllWithFilter(
			gosqlite.NewFilter().KeyHasNoBits(flagsField, permissionAdmin),
		)
		if err != nil {
			t.Fatalf("failed to select records without admin bit: %v", err)
		}
		assert.Len(t, noAdmin, 2, "expected records without the admin bit")
	})
}

func TestFilterNotIn(t *testing.T) {
	withTestDB(t, func(driver *gosqlite.Driver) {
		var (
			err     error
			handler *gosqlite.RegisteredStruct[Document]
		)

		if handler, err = gosqlite.Register(driver, Document{}); err != nil {
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

		filter := gosqlite.NewFilter().
			KeyCmp(handler.FieldByGoName("ID"), gosqlite.OpNotIn, []int{ids[1], ids[2]})

		results, err := handler.SelectAllWithFilter(filter)
		if err != nil {
			t.Fatalf("failed to select with not-in filter: %v", err)
		}

		assert.Len(t, results, 2, "expected two documents after not-in filter")
	})
}
