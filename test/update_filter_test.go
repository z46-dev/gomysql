package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/z46-dev/gomysql"
)

type Account struct {
	Username string `gomysql:"username,primary,unique"`
	Money    int    `gomysql:"money"`
}

func TestUpdateWithFilterReturning(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[Account]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(Account{}); err != nil {
		t.Fatalf("failed to register Account struct: %v", err)
	}

	accounts := []Account{
		{Username: "bob", Money: 100},
		{Username: "alice", Money: 30},
	}

	for i := range accounts {
		if err = handler.Insert(&accounts[i]); err != nil {
			t.Fatalf("failed to insert account %d: %v", i, err)
		}
	}

	filter := gomysql.NewFilter().
		KeyCmp(handler.FieldByGoName("Username"), gomysql.OpEqual, "bob").
		And().
		KeyCmp(handler.FieldByGoName("Money"), gomysql.OpGreaterThan, 50)

	returning := []*gomysql.RegisteredStructField{
		handler.FieldByGoName("Money"),
	}

	rows, err := handler.UpdateWithFilterReturning(
		filter,
		returning,
		gomysql.SetSub(handler.FieldByGoName("Money"), 25),
	)
	if err != nil {
		t.Fatalf("failed to update with returning: %v", err)
	}

	if assert.Len(t, rows, 1, "expected one returned row") {
		assert.EqualValues(t, 75, rows[0]["money"], "returned money should reflect updated value")
	}

	updated, err := handler.Select("bob")
	if err != nil {
		t.Fatalf("failed to select updated account: %v", err)
	}

	assert.EqualValues(t, 75, updated.Money, "money should be updated")
}

func TestUpdateWithFilter(t *testing.T) {
	var (
		err     error
		handler *gomysql.RegisteredStruct[Account]
	)

	if err = gomysql.Begin(":memory:"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	defer func() {
		if err := gomysql.Close(); err != nil {
			t.Fatalf("failed to close database connection: %v", err)
		}
	}()

	if handler, err = gomysql.Register(Account{}); err != nil {
		t.Fatalf("failed to register Account struct: %v", err)
	}

	accounts := []Account{
		{Username: "bob", Money: 10},
		{Username: "alice", Money: 40},
		{Username: "cora", Money: 5},
	}

	for i := range accounts {
		if err = handler.Insert(&accounts[i]); err != nil {
			t.Fatalf("failed to insert account %d: %v", i, err)
		}
	}

	filter := gomysql.NewFilter().
		KeyCmp(handler.FieldByGoName("Money"), gomysql.OpLessThan, 20)

	rows, err := handler.UpdateWithFilter(
		filter,
		gomysql.SetField(handler.FieldByGoName("Money"), 0),
	)
	if err != nil {
		t.Fatalf("failed to update with filter: %v", err)
	}

	assert.EqualValues(t, 2, rows, "expected two rows to update")

	bob, err := handler.Select("bob")
	if err != nil {
		t.Fatalf("failed to select bob: %v", err)
	}
	cora, err := handler.Select("cora")
	if err != nil {
		t.Fatalf("failed to select cora: %v", err)
	}
	alice, err := handler.Select("alice")
	if err != nil {
		t.Fatalf("failed to select alice: %v", err)
	}

	assert.EqualValues(t, 0, bob.Money, "bob should be updated")
	assert.EqualValues(t, 0, cora.Money, "cora should be updated")
	assert.EqualValues(t, 40, alice.Money, "alice should not be updated")
}
