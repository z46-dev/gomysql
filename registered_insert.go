package gomysql

import (
	"fmt"
	"reflect"
)

func (r *RegisteredStruct[T]) Insert(item *T) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}

	var (
		values []any
		elem   = reflect.ValueOf(item).Elem()
	)

	for _, field := range r.insertOrdered {
		if val, err := getSQLValueOf(field, elem.FieldByName(field.RealName)); err != nil {
			return fmt.Errorf("value conversion %s: %w", field.Opts.KeyName, err)
		} else {
			values = append(values, val)
		}
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	if result, err := r.db.db.Exec(r.insertSQL, values...); err != nil {
		return fmt.Errorf("insert fail %s: %w", r.Name, err)
	} else if field := r.PrimaryKeyField; field.Opts.PrimaryKey && field.Opts.AutoIncr {
		if lastInsertID, err := result.LastInsertId(); err != nil {
			return fmt.Errorf("auto-incr ID fail %s: %w", r.Name, err)
		} else {
			elem.FieldByName(field.RealName).SetInt(lastInsertID)
		}
	}

	return nil
}
