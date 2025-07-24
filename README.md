# gomysql package

When I make some of my projects, I use a MySQL database. I realized I ended up writing very similar code for each project to connect to the database, execute queries, and handle results. To avoid this repetition, I created the `gomysql` package. It utilizes go's `reflect` package to handle different types of structs to create tables, and execute a basic insert, select, update, delete, and list operations.

# Usage

To use the `gomysql` package, you need to import it in your Go project. Here's a basic example of how to use it:

```go
import "github.com/z46-dev/gomysql"
```

Now, let's create a struct that represents a table in the database. This package uses custom struct tags to define table structure, and its fields. Here's an example of a struct that represents a `User` table:

```go
type User struct {
	ID       int       `gomysql:"id,primary,increment"`
    Username string    `gomysql:"username,unique"`
}
```