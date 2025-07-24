package gomysql

import (
	"fmt"
	"strings"
)

// CREATE TABLE IF NOT EXISTS X (key1 INTEGER PRIMARY KEY, key2 TEXT, ...);
// with autoincrement: INSERT OR REPLACE INTO X (key2, ...) VALUES (?, ...);
// without autoincrement: INSERT OR REPLACE INTO X (key1, key2, ...) VALUES (?, ?, ...);
// SELECT key2, ... FROM X WHERE key1 = ?;
// UPDATE X SET key2 = ?, ... WHERE key1 = ?;
// DELETE FROM X WHERE key1 = ?;
// SELECT key1 FROM X;
func generateSQLStatements[T any](r *RegisteredStruct[T]) {
	var pKey RegisteredStructField
	for _, field := range r.Fields {
		if field.Opts.PrimaryKey {
			pKey = field
			break
		}
	}

	r.createTableSQL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", r.Name)
	var pKeyStr = fmt.Sprintf("%s %s PRIMARY KEY", pKey.Opts.KeyName, typeNameString(pKey.InternalType))

	if pKey.Opts.AutoIncr {
		pKeyStr += " AUTOINCREMENT"
	} else {
		r.insertOrdered = append(r.insertOrdered, pKey)
	}

	if pKey.Opts.Unique {
		pKeyStr += " UNIQUE"
	}

	if pKey.Opts.NotNull {
		pKeyStr += " NOT NULL"
	}

	r.createTableSQL += pKeyStr + ", "

	for _, field := range r.Fields {
		if field.Opts.PrimaryKey {
			continue
		}

		var str = fmt.Sprintf("%s %s", field.Opts.KeyName, typeNameString(field.InternalType))
		if field.Opts.Unique {
			str += " UNIQUE"
		}

		if field.Opts.NotNull {
			str += " NOT NULL"
		}

		r.createTableSQL += str + ", "
		r.insertOrdered = append(r.insertOrdered, field)
		r.nonInsertionOrdered = append(r.nonInsertionOrdered, field)
	}

	var mapper = func(fields []RegisteredStructField, joiner string) string {
		var parts []string
		for _, field := range fields {
			parts = append(parts, field.Opts.KeyName)
		}
		return strings.Join(parts, joiner)
	}

	r.createTableSQL = strings.TrimSuffix(r.createTableSQL, ", ") + ");"
	r.insertSQL = fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES (%s);", r.Name, mapper(r.insertOrdered, ", "), strings.Repeat("?, ", len(r.insertOrdered)-1)+"?")
	r.updateSQL = fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?;", r.Name, mapper(r.nonInsertionOrdered, " = ?, ")+" = ?", pKey.Opts.KeyName)
	r.selectSQL = fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?;", mapper(r.nonInsertionOrdered, ", "), r.Name, pKey.Opts.KeyName)
	r.deleteSQL = fmt.Sprintf("DELETE FROM %s WHERE %s = ?;", r.Name, pKey.Opts.KeyName)
	r.listSQL = fmt.Sprintf("SELECT %s FROM %s;", pKey.Opts.KeyName, r.Name)
	r.selectAllSQL = fmt.Sprintf("SELECT %s, %s FROM %s;", pKey.Opts.KeyName, mapper(r.nonInsertionOrdered, ", "), r.Name)
	r.PrimaryKeyField = pKey
}
