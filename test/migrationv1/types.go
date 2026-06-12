package migrationv1

type AddItem struct {
	ID   int    `gosqlite:"id,primary,increment"`
	Name string `gosqlite:"name"`
}

type DropItem struct {
	ID   int    `gosqlite:"id,primary,increment"`
	Name string `gosqlite:"name"`
	Age  int    `gosqlite:"age"`
}

type TypeItem struct {
	ID     int `gosqlite:"id,primary,increment"`
	Active int `gosqlite:"active"`
}

type Parent struct {
	ID   int    `gosqlite:"id,primary,increment"`
	Name string `gosqlite:"name"`
}

type Child struct {
	ID       int `gosqlite:"id,primary,increment"`
	ParentID int `gosqlite:"parent_id"`
}

type CascadeChild struct {
	ID       int `gosqlite:"id,primary,increment"`
	ParentID int `gosqlite:"parent_id,fkey:Parent.id"`
}
