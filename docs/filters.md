# Filters and pagination

Filters build SQL WHERE clauses with placeholders and arguments.

Supported comparison operators:

- `gomysql.OpEqual`
- `gomysql.OpNotEqual`
- `gomysql.OpGreaterThan`
- `gomysql.OpLessThan`
- `gomysql.OpGreaterThanOrEqual`
- `gomysql.OpLessThanOrEqual`
- `gomysql.OpLike`
- `gomysql.OpIn`
- `gomysql.OpNotIn`
- `gomysql.OpIsNull`
- `gomysql.OpIsNotNull`

```go
filter := gomysql.NewFilter().
	KeyCmp(handler.FieldByGoName("Title"), gomysql.OpLike, "%report%").
	And().
	KeyCmp(handler.FieldByGoName("Published"), gomysql.OpEqual, true)

docs, err := handler.SelectAllWithFilter(filter)
```

## Grouping

```go
filter := gomysql.NewFilter().
	OpenGroup().
	KeyCmp(handler.FieldByGoName("Status"), gomysql.OpEqual, "draft").
	Or().
	KeyCmp(handler.FieldByGoName("Status"), gomysql.OpEqual, "review").
	CloseGroup().
	And().
	KeyCmp(handler.FieldByGoName("Owner"), gomysql.OpEqual, "alice")
```

## IN and NOT IN

```go
filter := gomysql.NewFilter().
	KeyCmp(handler.FieldByGoName("ID"), gomysql.OpIn, []int{1, 2, 3})
```

## Ordering, limit, and offset

```go
filter := gomysql.NewFilter().
	KeyCmp(handler.FieldByGoName("Score"), gomysql.OpGreaterThanOrEqual, 10).
	Ordering(handler.FieldByGoName("Score"), false).
	Limit(50).
	Offset(100)
```

## Compare `time.Time` fields

`time.Time` fields are stored as SQL `DATETIME` values, so range filters and ordering work natively in SQL.

```go
cutoff := time.Now().UTC().Add(-24 * time.Hour)

olderDocs, err := handler.SelectAllWithFilter(
	gomysql.NewFilter().
		KeyCmp(handler.FieldByGoName("Creation"), gomysql.OpLessThan, cutoff),
)
if err != nil {
	panic(err)
}

recentDocs, err := handler.SelectAllWithFilter(
	gomysql.NewFilter().
		KeyCmp(handler.FieldByGoName("Creation"), gomysql.OpGreaterThanOrEqual, cutoff).
		Ordering(handler.FieldByGoName("Creation"), false),
)
if err != nil {
	panic(err)
}

_ = olderDocs
_ = recentDocs
```

Use `UTC()` consistently when writing or filtering timestamps so comparisons stay predictable.

## Count rows without loading them

```go
count, err := handler.CountWithFilter(
	gomysql.NewFilter().
		KeyCmp(handler.FieldByGoName("Published"), gomysql.OpEqual, true),
)
if err != nil {
	panic(err)
}
```

## Delete the oldest rows quickly

```go
deleted, err := handler.DeleteWithFilter(
	gomysql.NewFilter().
		Ordering(handler.FieldByGoName("Creation"), true).
		Limit(100),
)
if err != nil {
	panic(err)
}
_ = deleted
```

`DeleteWithFilter` uses the primary key under the hood, so ordered and limited deletes work for cases like pruning the oldest rows from a large table.
