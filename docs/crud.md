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
