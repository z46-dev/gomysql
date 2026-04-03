# CRUD operations

## Insert

```go
doc := &Document{Title: "Hello", Body: "World"}
if err := handler.Insert(doc); err != nil {
	panic(err)
}
```

## Select by primary key

```go
doc, err := handler.Select(1)
if err != nil {
	panic(err)
}
```

## Update by primary key

```go
doc.Title = "Updated"
if err := handler.Update(doc); err != nil {
	panic(err)
}
```

## Delete by primary key

```go
if err := handler.Delete(1); err != nil {
	panic(err)
}
```

## Count rows

```go
total, err := handler.Count()
if err != nil {
	panic(err)
}
```

## Count rows with a filter

```go
published, err := handler.CountWithFilter(
	gomysql.NewFilter().
		KeyCmp(handler.FieldByGoName("Published"), gomysql.OpEqual, true),
)
if err != nil {
	panic(err)
}
```

## Delete rows with a filter

```go
deleted, err := handler.DeleteWithFilter(
	gomysql.NewFilter().
		KeyCmp(handler.FieldByGoName("Published"), gomysql.OpEqual, false),
)
if err != nil {
	panic(err)
}
_ = deleted
```

## List primary keys

```go
ids, err := handler.List()
if err != nil {
	panic(err)
}
```

## Select all rows

```go
docs, err := handler.SelectAll()
if err != nil {
	panic(err)
}
```
