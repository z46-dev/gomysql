# Filters and pagination

Filters build SQL WHERE clauses with placeholders and arguments.

Supported comparison operators:

- `gosqlite.OpEqual`
- `gosqlite.OpNotEqual`
- `gosqlite.OpGreaterThan`
- `gosqlite.OpLessThan`
- `gosqlite.OpGreaterThanOrEqual`
- `gosqlite.OpLessThanOrEqual`
- `gosqlite.OpLike`
- `gosqlite.OpIn`
- `gosqlite.OpNotIn`
- `gosqlite.OpIsNull`
- `gosqlite.OpIsNotNull`

```go
filter := gosqlite.NewFilter().
	KeyCmp(handler.FieldByGoName("Title"), gosqlite.OpLike, "%report%").
	And().
	KeyCmp(handler.FieldByGoName("Published"), gosqlite.OpEqual, true)

docs, err := handler.SelectAllWithFilter(filter)
```

## Grouping

```go
filter := gosqlite.NewFilter().
	OpenGroup().
	KeyCmp(handler.FieldByGoName("Status"), gosqlite.OpEqual, "draft").
	Or().
	KeyCmp(handler.FieldByGoName("Status"), gosqlite.OpEqual, "review").
	CloseGroup().
	And().
	KeyCmp(handler.FieldByGoName("Owner"), gosqlite.OpEqual, "alice")
```

## IN and NOT IN

```go
filter := gosqlite.NewFilter().
	KeyCmp(handler.FieldByGoName("ID"), gosqlite.OpIn, []int{1, 2, 3})
```

## Bitwise integer filters

Integer fields can be filtered with SQLite bitwise `&` predicates:

```go
const (
	PermissionRead = 1 << iota
	PermissionWrite
	PermissionAdmin
)

filter := gosqlite.NewFilter().
	KeyHasAllBits(handler.FieldByGoName("Flags"), PermissionRead|PermissionWrite)
```

- `KeyHasAnyBits(field, mask)` matches rows where `(field & mask) != 0`.
- `KeyHasAllBits(field, mask)` matches rows where `(field & mask) = mask`.
- `KeyHasNoBits(field, mask)` matches rows where `(field & mask) = 0`.

## Ordering, limit, and offset

```go
filter := gosqlite.NewFilter().
	KeyCmp(handler.FieldByGoName("Score"), gosqlite.OpGreaterThanOrEqual, 10).
	Ordering(handler.FieldByGoName("Score"), false).
	Limit(50).
	Offset(100)
```

## Compare `time.Time` fields

`time.Time` fields are stored as SQL `DATETIME` values, so range filters and ordering work natively in SQL.

```go
cutoff := time.Now().UTC().Add(-24 * time.Hour)

olderDocs, err := handler.SelectAllWithFilter(
	gosqlite.NewFilter().
		KeyCmp(handler.FieldByGoName("Creation"), gosqlite.OpLessThan, cutoff),
)
if err != nil {
	panic(err)
}

recentDocs, err := handler.SelectAllWithFilter(
	gosqlite.NewFilter().
		KeyCmp(handler.FieldByGoName("Creation"), gosqlite.OpGreaterThanOrEqual, cutoff).
		Ordering(handler.FieldByGoName("Creation"), false),
)
if err != nil {
	panic(err)
}

_ = olderDocs
_ = recentDocs
```

Use `UTC()` consistently when writing or filtering timestamps so comparisons stay predictable.

## Compare `time.Duration` fields

`time.Duration` fields are stored as SQL `INTEGER` nanoseconds, so numeric range filters and ordering also work natively in SQL.

```go
slowJobs, err := handler.SelectAllWithFilter(
	gosqlite.NewFilter().
		KeyCmp(handler.FieldByGoName("Timeout"), gosqlite.OpGreaterThanOrEqual, 30*time.Second).
		Ordering(handler.FieldByGoName("Timeout"), false),
)
if err != nil {
	panic(err)
}

_ = slowJobs
```

Durations round-trip as `time.Duration`, but if you inspect the raw SQL column directly you will see nanoseconds.

## Count rows without loading them

```go
count, err := handler.CountWithFilter(
	gosqlite.NewFilter().
		KeyCmp(handler.FieldByGoName("Published"), gosqlite.OpEqual, true),
)
if err != nil {
	panic(err)
}
```

## Delete the oldest rows quickly

```go
deleted, err := handler.DeleteWithFilter(
	gosqlite.NewFilter().
		Ordering(handler.FieldByGoName("Creation"), true).
		Limit(100),
)
if err != nil {
	panic(err)
}
_ = deleted
```

`DeleteWithFilter` uses the primary key under the hood, so ordered and limited deletes work for cases like pruning the oldest rows from a large table.
