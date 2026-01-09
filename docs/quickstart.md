# Quickstart

## Install

```go
import "github.com/z46-dev/gomysql"
```

## Define a struct

```go
type User struct {
	ID       int    `gomysql:"id,primary,increment"`
	Username string `gomysql:"username,unique"`
	Email    string `gomysql:"email"`
}
```

## Connect and register

```go
if err := gomysql.Begin(":memory:"); err != nil {
	panic(err)
}
defer gomysql.Close()

users, err := gomysql.Register(User{})
if err != nil {
	panic(err)
}
```

## Basic CRUD

```go
user := &User{Username: "alice", Email: "a@example.com"}
if err := users.Insert(user); err != nil {
	panic(err)
}

fetched, err := users.Select(user.ID)
if err != nil {
	panic(err)
}

fetched.Email = "new@example.com"
if err := users.Update(fetched); err != nil {
	panic(err)
}

if err := users.Delete(fetched.ID); err != nil {
	panic(err)
}
```
