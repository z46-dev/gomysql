package migrationv1

type AddItem struct {
	ID   int    `gomysql:"id,primary,increment"`
	Name string `gomysql:"name"`
}

type DropItem struct {
	ID   int    `gomysql:"id,primary,increment"`
	Name string `gomysql:"name"`
	Age  int    `gomysql:"age"`
}

type TypeItem struct {
	ID     int `gomysql:"id,primary,increment"`
	Active int `gomysql:"active"`
}
