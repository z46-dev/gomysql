# Update expressions and RETURNING

Use update assignments to build expressions without a full struct update.

## Basic update with filter

```go
filter := gosqlite.NewFilter().
	KeyCmp(handler.FieldByGoName("Username"), gosqlite.OpEqual, "bob").
	And().
	KeyCmp(handler.FieldByGoName("Money"), gosqlite.OpGreaterThan, 20)

rows, err := handler.UpdateWithFilter(
	filter,
	gosqlite.SetSub(handler.FieldByGoName("Money"), 10),
)
```

## Set a literal value

```go
_, err := handler.UpdateWithFilter(
	gosqlite.NewFilter().KeyCmp(handler.FieldByGoName("Status"), gosqlite.OpEqual, "draft"),
	gosqlite.SetField(handler.FieldByGoName("Status"), "archived"),
)
```

## Custom expressions

```go
_, err := handler.UpdateWithFilter(
	gosqlite.NewFilter().KeyCmp(handler.FieldByGoName("ID"), gosqlite.OpEqual, 42),
	gosqlite.SetExpr(handler.FieldByGoName("Score"), "MAX(Score, ?)", 100),
)
```

## RETURNING

```go
rows, err := handler.UpdateWithFilterReturning(
	filter,
	[]*gosqlite.RegisteredStructField{
		handler.FieldByGoName("Money"),
	},
	gosqlite.SetSub(handler.FieldByGoName("Money"), 10),
)

updatedMoney := rows[0]["money"]
```

Returned values are keyed by SQL column name from the struct tag.
