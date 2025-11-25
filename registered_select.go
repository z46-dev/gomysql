package gomysql

import (
	"bytes"
	"database/sql"
	"encoding/gob"
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

		switch field.InternalType {
		case TypeRepInt:
			elem.FieldByName(field.RealName).SetInt(raw.(int64))
		case TypeRepUint:
			elem.FieldByName(field.RealName).SetUint(raw.(uint64))
		case TypeRepString:
			elem.FieldByName(field.RealName).SetString(raw.(string))
		case TypeRepBool:
			boolean, err := sqlBool(raw)
			if err != nil {
				return nil, fmt.Errorf("convert %s to bool: %w", field.Opts.KeyName, err)
			}
			elem.FieldByName(field.RealName).SetBool(boolean)
		case TypeRepArrayBlob, TypeRepStructBlob, TypeRepMapBlob:
			target := reflect.New(field.Type).Interface()
			if err := gob.NewDecoder(bytes.NewReader(raw.([]byte))).Decode(target); err != nil {
				return nil, fmt.Errorf("gob decode %s: %w", field.Opts.KeyName, err)
			}

			elem.FieldByName(field.RealName).Set(reflect.ValueOf(target).Elem())
		case TypeRepFloat:
			elem.FieldByName(field.RealName).SetFloat(raw.(float64))
		case TypeRepPointer:
			if raw == nil {
				elem.FieldByName(field.RealName).Set(reflect.Zero(field.Type))
				continue
			}
			target := reflect.New(field.Type.Elem())
			if err := gob.NewDecoder(bytes.NewReader(raw.([]byte))).Decode(target.Interface()); err != nil {
				return nil, fmt.Errorf("gob decode %s: %w", field.Opts.KeyName, err)
			}
			elem.FieldByName(field.RealName).Set(target)
		default:
			return nil, fmt.Errorf("unsupported type for field %s", field.Opts.KeyName)
		}
	}

	if primaryKeyField := r.PrimaryKeyField; primaryKeyField.RealName != "" {
		elem.FieldByName(primaryKeyField.RealName).Set(reflect.ValueOf(primaryKeyValue))
	}

	return item, nil
}
