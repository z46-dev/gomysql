package gomysql

import (
	"database/sql"
	"fmt"
	"reflect"
)

func (r *RegisteredStruct[T]) Select(primaryKeyValue any) (item *T, err error) {
	if r.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	item = new(T)

	var (
		values   []any = make([]any, len(r.nonInsertionOrdered))
		scanArgs []any = make([]any, len(r.nonInsertionOrdered))
		elem           = reflect.ValueOf(item).Elem()
	)

	for i := range r.nonInsertionOrdered {
		scanArgs[i] = &values[i]
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	row := r.db.db.QueryRow(r.selectSQL, primaryKeyValue)
	if err = row.Scan(scanArgs...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to select from %s: %w", r.Name, err)
	}

	for i, field := range r.nonInsertionOrdered {
		raw := values[i]
		fieldValue := elem.FieldByIndex(field.Index)
		if err := assignDecodedValue(fieldValue, field, raw); err != nil {
			return nil, err
		}
	}

	if primaryKeyField := r.PrimaryKeyField; primaryKeyField.RealName != "" {
		fieldValue := elem.FieldByIndex(primaryKeyField.Index)
		if fieldValue.Kind() == reflect.Pointer {
			target := reflect.New(fieldValue.Type().Elem())
			target.Elem().Set(reflect.ValueOf(primaryKeyValue).Convert(fieldValue.Type().Elem()))
			fieldValue.Set(target)
		} else {
			fieldValue.Set(reflect.ValueOf(primaryKeyValue).Convert(fieldValue.Type()))
		}
	}

	return item, nil
}
