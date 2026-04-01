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

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	for _, field := range r.insertOrdered {
		fieldValue := elem.FieldByIndex(field.Index)
		if field.Opts.AutoIncr && !field.Opts.PrimaryKey && fieldValue.IsZero() {
			nextValue, err := r.nextAutoIncrementValue(field)
			if err != nil {
				return err
			}
			target := fieldValue
			if fieldValue.Kind() == reflect.Pointer {
				target = reflect.New(fieldValue.Type().Elem()).Elem()
			}
			switch target.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				target.SetInt(nextValue)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				target.SetUint(uint64(nextValue))
			default:
				return fmt.Errorf("auto-increment unsupported type %s for %s", target.Kind(), field.Opts.KeyName)
			}
			if fieldValue.Kind() == reflect.Pointer {
				ptr := reflect.New(fieldValue.Type().Elem())
				ptr.Elem().Set(target)
				fieldValue.Set(ptr)
			}
		}

		if val, err := getSQLValueOf(field, fieldValue); err != nil {
			return fmt.Errorf("value conversion %s: %w", field.Opts.KeyName, err)
		} else {
			values = append(values, val)
		}
	}

	if result, err := r.db.db.Exec(r.insertSQL, values...); err != nil {
		return fmt.Errorf("insert fail %s: %w", r.Name, err)
	} else if field := r.PrimaryKeyField; field.Opts.PrimaryKey && field.Opts.AutoIncr {
		if lastInsertID, err := result.LastInsertId(); err != nil {
			return fmt.Errorf("auto-incr ID fail %s: %w", r.Name, err)
		} else {
			fieldValue := elem.FieldByIndex(field.Index)
			if fieldValue.Kind() == reflect.Pointer {
				ptr := reflect.New(fieldValue.Type().Elem())
				ptr.Elem().SetInt(lastInsertID)
				fieldValue.Set(ptr)
			} else {
				fieldValue.SetInt(lastInsertID)
			}
		}
	}

	return nil
}

func (r *RegisteredStruct[T]) nextAutoIncrementValue(field RegisteredStructField) (int64, error) {
	query := fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) + 1 FROM %s;", field.Opts.KeyName, r.Name)
	var next int64
	if err := r.db.db.QueryRow(query).Scan(&next); err != nil {
		return 0, fmt.Errorf("auto-increment query %s: %w", field.Opts.KeyName, err)
	}
	return next, nil
}
