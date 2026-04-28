# Struct tags and types

## Tags

The `gomysql` struct tag defines the SQL column name and options.

```go
type Document struct {
	ID       int    `gomysql:"id,primary,increment"`
	Title    string `gomysql:"title,unique"`
	Body     string `gomysql:"body"`
	IsPublic bool   `gomysql:"is_public"`
}
```

Supported options:

- `primary` marks the primary key (one per struct).
- `increment` enables autoincrement on the primary key.
- `unique` adds a UNIQUE constraint.
- `notnull` adds a NOT NULL constraint.
- `fkey:StructGoName.mysqlFieldName` adds a foreign key reference to another registered table.
- `ondelete:cascade` adds `ON DELETE CASCADE` to the field's foreign key.

Example:

```go
type User struct {
	ID int `gomysql:"id,primary,increment"`
}

type Session struct {
	ID     int `gomysql:"id,primary,increment"`
	UserID int `gomysql:"user_id,fkey:User.id"`
}
```

To delete child rows automatically when the parent row is removed:

```go
type AuditLog struct {
	ID     int `gomysql:"id,primary,increment"`
	UserID int `gomysql:"user_id,fkey:User.id,ondelete:cascade"`
}
```

If you omit `ondelete:cascade`, SQLite will reject parent deletes while child rows still reference them.

## Supported field kinds

- Integers (signed/unsigned)
- Strings
- Bools
- Float32/Float64
- Arrays/slices (stored as gob-encoded blobs)
- Structs (stored as gob-encoded blobs)
- Maps (stored as gob-encoded blobs)
- Pointers to structs (stored as gob-encoded blobs, nullable)

## Field lookup helpers

Use these to reference a column when building filters or update expressions:

```go
field := handler.FieldByGoName("Title")
sqlField := handler.FieldBySQLName("title")
```
