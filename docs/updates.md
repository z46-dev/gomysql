# Update expressions and RETURNING

Use update assignments to build expressions without a full struct update.

## Basic update with filter

```go
filter := gomysql.NewFilter().
	KeyCmp(handler.FieldByGoName("Username"), gomysql.OpEqual, "bob").
	And().
	KeyCmp(handler.FieldByGoName("Money"), gomysql.OpGreaterThan, 20)

rows, err := handler.UpdateWithFilter(
	filter,
	gomysql.SetSub(handler.FieldByGoName("Money"), 10),
)
```

## Set a literal value

```go
_, err := handler.UpdateWithFilter(
	gomysql.NewFilter().KeyCmp(handler.FieldByGoName("Status"), gomysql.OpEqual, "draft"),
	gomysql.SetField(handler.FieldByGoName("Status"), "archived"),
)
```

## Custom expressions

```go
_, err := handler.UpdateWithFilter(
	gomysql.NewFilter().KeyCmp(handler.FieldByGoName("ID"), gomysql.OpEqual, 42),
	gomysql.SetExpr(handler.FieldByGoName("Score"), "MAX(Score, ?)", 100),
)
```

## RETURNING

```go
rows, err := handler.UpdateWithFilterReturning(
	filter,
	[]*gomysql.RegisteredStructField{
		handler.FieldByGoName("Money"),
	},
	gomysql.SetSub(handler.FieldByGoName("Money"), 10),
)

updatedMoney := rows[0]["money"]
```

Returned values are keyed by SQL column name from the struct tag.
