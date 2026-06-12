package migrationv2

type AddItem struct {
	ID   int    `gosqlite:"id,primary,increment"`
	Name string `gosqlite:"name"`
	Age  int    `gosqlite:"age"`
}

type DropItem struct {
	ID   int    `gosqlite:"id,primary,increment"`
	Name string `gosqlite:"name"`
}

type TypeItem struct {
	ID     int  `gosqlite:"id,primary,increment"`
	Active bool `gosqlite:"active"`
}

type Parent struct {
	ID   int    `gosqlite:"id,primary,increment"`
	Name string `gosqlite:"name"`
}

type Child struct {
	ID       int `gosqlite:"id,primary,increment"`
	ParentID int `gosqlite:"parent_id,fkey:Parent.id"`
}

type CascadeChild struct {
	ID       int `gosqlite:"id,primary,increment"`
	ParentID int `gosqlite:"parent_id,fkey:Parent.id,ondelete:cascade"`
}
