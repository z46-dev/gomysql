# Filters and pagination

Filters build SQL WHERE clauses with placeholders and arguments.

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
