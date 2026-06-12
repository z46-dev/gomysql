# gosqlite package

When I make some of my projects, I use a SQLite database. I realized I ended up writing very similar code for each project to connect to the database, execute queries, and handle results. To avoid this repetition, I created the `gosqlite` package. It utilizes go's `reflect` package to handle different types of structs to create tables, and execute a basic insert, select, update, delete, and list operations.

# Usage

To use the `gosqlite` package, you need to import it in your Go project. Here's a basic example of how to use it:

```go
import "github.com/z46-dev/gosqlite"
```

Now, let's create a struct that represents a table in the database. This package uses custom struct tags to define table structure, and its fields. Here's an example of a struct that represents a `User` table:

```go
type User struct {
    Username string     `gosqlite:"username,primary,unique"`
    Password string     `gosqlite:"password"`
    Email    string     `gosqlite:"email,unique"`
    CreatedAt time.Time `gosqlite:"creation"`
}
```

Foreign keys are declared in the same tag with `fkey:StructGoName.SQLiteFieldName`. Add `ondelete:cascade` to opt into cascading child deletes:

```go
type Team struct {
    ID int `gosqlite:"id,primary,increment"`
}

type Player struct {
    ID     int `gosqlite:"id,primary,increment"`
    TeamID int `gosqlite:"team_id,fkey:Team.id"`
}

type Match struct {
    ID     int `gosqlite:"id,primary,increment"`
    TeamID int `gosqlite:"team_id,fkey:Team.id,ondelete:cascade"`
}
```

Without `ondelete:cascade`, deleting a parent row while children still reference it will fail under SQLite foreign key enforcement.

Now, we should register this struct to set up the database tables:

```go
func main() {
    var (
        handle *gosqlite.RegisteredStruct[User]
        err    error
    )

    if err = gosqlite.Begin(":memory:"); err != nil {
        panic(err)
    }

    if handle, err = gosqlite.Register(User{}); err != nil {
        panic(err)
    }
}
```

You can now call functions on the `handle` to perform database operations.

`time.Time` fields are stored as SQL `DATETIME` values, and `time.Duration` fields are stored as SQL `INTEGER` nanoseconds, so both can be filtered and ordered directly in SQL.

For larger tables, filters can also be used for efficient row counts and pruning the oldest rows without loading full records:

```go
total, err := handle.Count()
if err != nil {
    panic(err)
}

deleted, err := handle.DeleteWithFilter(
    gosqlite.NewFilter().
        Ordering(handle.FieldByGoName("CreatedAt"), true).
        Limit(100),
)
if err != nil {
    panic(err)
}

_ = total
_ = deleted
```

## Contributing

This was thrown together in a few hours, so there are likely many improvements that can be made. If you have suggestions or improvements, feel free to open an issue or submit a pull request. I would love to see this package grow.
