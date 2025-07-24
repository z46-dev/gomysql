package gomysql

import (
	"fmt"
	"reflect"
)

func (r *RegisteredStruct[T]) Update(item *T) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}

	var (
		values []any
		elem   = reflect.ValueOf(item).Elem()
	)

	for _, field := range r.nonInsertionOrdered {
		if val, err := getSQLValueOf(field, elem.FieldByName(field.RealName)); err != nil {
			return fmt.Errorf("value conversion %s: %w", field.Opts.KeyName, err)
		} else {
			values = append(values, val)
		}
	}

	if primaryKeyField := r.PrimaryKeyField; primaryKeyField.RealName != "" {
		values = append(values, elem.FieldByName(primaryKeyField.RealName).Interface())
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	if _, err := r.db.db.Exec(r.updateSQL, values...); err != nil {
		return fmt.Errorf("update fail %s: %w", r.Name, err)
	}

	return nil
}
