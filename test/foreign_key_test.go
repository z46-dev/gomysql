package test

import (
	"testing"

	"github.com/z46-dev/gosqlite"
)

type RestrictParent struct {
	ID   int    `gosqlite:"id,primary,increment"`
	Name string `gosqlite:"name"`
}

type RestrictChild struct {
	ID       int `gosqlite:"id,primary,increment"`
	ParentID int `gosqlite:"parent_id,fkey:RestrictParent.id"`
}

type CascadeParent struct {
	ID   int    `gosqlite:"id,primary,increment"`
	Name string `gosqlite:"name"`
}

type CascadeChild struct {
	ID       int `gosqlite:"id,primary,increment"`
	ParentID int `gosqlite:"parent_id,fkey:CascadeParent.id,ondelete:cascade"`
}

func TestForeignKeyDeleteRestrictedByDefault(t *testing.T) {
	withTestDB(t, func(driver *gosqlite.Driver) {
		parentHandler, err := gosqlite.Register(driver, RestrictParent{})
		if err != nil {
			t.Fatalf("failed to register parent struct: %v", err)
		}

		childHandler, err := gosqlite.Register(driver, RestrictChild{})
		if err != nil {
			t.Fatalf("failed to register child struct: %v", err)
		}

		parent := &RestrictParent{Name: "parent"}
		if err := parentHandler.Insert(parent); err != nil {
			t.Fatalf("failed to insert parent: %v", err)
		}

		child := &RestrictChild{ParentID: parent.ID}
		if err := childHandler.Insert(child); err != nil {
			t.Fatalf("failed to insert child: %v", err)
		}

		if err := parentHandler.Delete(parent.ID); err == nil {
			t.Fatalf("expected parent delete to fail without ON DELETE CASCADE")
		}

		if got, err := childHandler.Select(child.ID); err != nil {
			t.Fatalf("failed to select child after blocked parent delete: %v", err)
		} else if got == nil {
			t.Fatalf("expected child to remain after blocked parent delete")
		}
	})
}

func TestForeignKeyCascadeDelete(t *testing.T) {
	withTestDB(t, func(driver *gosqlite.Driver) {
		parentHandler, err := gosqlite.Register(driver, CascadeParent{})
		if err != nil {
			t.Fatalf("failed to register parent struct: %v", err)
		}

		childHandler, err := gosqlite.Register(driver, CascadeChild{})
		if err != nil {
			t.Fatalf("failed to register child struct: %v", err)
		}

		parent := &CascadeParent{Name: "parent"}
		if err := parentHandler.Insert(parent); err != nil {
			t.Fatalf("failed to insert parent: %v", err)
		}

		child := &CascadeChild{ParentID: parent.ID}
		if err := childHandler.Insert(child); err != nil {
			t.Fatalf("failed to insert child: %v", err)
		}

		if err := parentHandler.Delete(parent.ID); err != nil {
			t.Fatalf("failed to delete parent with ON DELETE CASCADE: %v", err)
		}

		if got, err := childHandler.Select(child.ID); err != nil {
			t.Fatalf("failed to select child after cascaded parent delete: %v", err)
		} else if got != nil {
			t.Fatalf("expected child to be deleted by cascade")
		}
	})
}
