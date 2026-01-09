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
