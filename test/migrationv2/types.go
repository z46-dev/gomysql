package migrationv2

type AddItem struct {
	ID   int    `gomysql:"id,primary,increment"`
	Name string `gomysql:"name"`
	Age  int    `gomysql:"age"`
}

type DropItem struct {
	ID   int    `gomysql:"id,primary,increment"`
	Name string `gomysql:"name"`
}

type TypeItem struct {
	ID     int  `gomysql:"id,primary,increment"`
	Active bool `gomysql:"active"`
}
