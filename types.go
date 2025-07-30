package gomysql

import (
	"fmt"
	"reflect"
)

type RegisteredStructField struct {
	Opts         SQLTagOpts
	RealName     string
	Type         reflect.Type
	InternalType TypeRepresentation
}

type RegisteredStruct[T any] struct {
	db                                                                                *Driver
	Name                                                                              string
	Type                                                                              reflect.Type
	Fields                                                                            []RegisteredStructField
	createTableSQL, insertSQL, selectSQL, updateSQL, deleteSQL, listSQL, selectAllSQL string
	PrimaryKeyField                                                                   RegisteredStructField
	insertOrdered, nonInsertionOrdered                                                []RegisteredStructField
}

type TypeRepresentation uint8

const (
	TypeRepInt TypeRepresentation = iota
	TypeRepUint
	TypeRepString
	TypeRepBool
	TypeRepArrayBlob
	TypeRepStructBlob
	TypeRepFloat
	TypeRepMapBlob
	TypeRepPointer
)

func typeNameString(t TypeRepresentation) string {
	switch t {
	case TypeRepInt:
		return "INTEGER"
	case TypeRepUint:
		return "INTEGER UNSIGNED"
	case TypeRepString:
		return "TEXT"
	case TypeRepBool:
		return "BOOLEAN"
	case TypeRepArrayBlob, TypeRepMapBlob, TypeRepStructBlob, TypeRepPointer:
		return "BLOB"
	case TypeRepFloat:
		return "FLOAT"
	default:
		return "UNKNOWN"
	}
}

var (
	ErrDatabaseInitialized    = fmt.Errorf("database already initialized")
	ErrDatabaseNotInitialized = fmt.Errorf("database not initialized")
)

type SQLOperator string

const (
	OpEqual              SQLOperator = "="
	OpNotEqual           SQLOperator = "!="
	OpGreaterThan        SQLOperator = ">"
	OpLessThan           SQLOperator = "<"
	OpGreaterThanOrEqual SQLOperator = ">="
	OpLessThanOrEqual    SQLOperator = "<="
	OpLike               SQLOperator = "LIKE"
	OpIn                 SQLOperator = "IN"
	OpNotIn              SQLOperator = "NOT IN"
	OpIsNull             SQLOperator = "IS NULL"
	OpIsNotNull          SQLOperator = "IS NOT NULL"
)

type Filter struct {
	args                                     []any
	whereTokens                              []string
	orderByClause, limitClause, offsetClause string
	lastWasJoiner                            bool
}
