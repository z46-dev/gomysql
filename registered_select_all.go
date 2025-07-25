package gomysql

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"
)

func (r *RegisteredStruct[T]) selectAll(sql string, args ...any) ([]*T, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	rows, err := r.db.db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all from %s: %w", r.Name, err)
	}
	defer rows.Close()

	var results []*T

	for rows.Next() {
		item := new(T)

		values := make([]any, len(r.nonInsertionOrdered)+1)
		scanArgs := make([]any, len(r.nonInsertionOrdered)+1)
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan all from %s: %w", r.Name, err)
		}

		elem := reflect.ValueOf(item).Elem()
		pkValue := values[0]
		if pkValue == nil {
			return nil, fmt.Errorf("primary key value is nil for %s", r.Name)
		}

		pkField := elem.FieldByName(r.PrimaryKeyField.RealName)
		switch pkField.Kind() {
		case reflect.Int, reflect.Int64:
			pkField.SetInt(pkValue.(int64))
		case reflect.String:
			pkField.SetString(pkValue.(string))
		default:
			return nil, fmt.Errorf("unsupported primary key type %s for field %s", pkField.Kind(), r.PrimaryKeyField.RealName)
		}

		for i, field := range r.nonInsertionOrdered {
			raw := values[i+1]
			switch field.InternalType {
			case TypeRepInt:
				elem.FieldByName(field.RealName).SetInt(raw.(int64))
			case TypeRepUint:
				elem.FieldByName(field.RealName).SetUint(raw.(uint64))
			case TypeRepString:
				elem.FieldByName(field.RealName).SetString(raw.(string))
			case TypeRepBool:
				elem.FieldByName(field.RealName).SetBool(raw.(bool))
			case TypeRepArrayBlob, TypeRepStructBlob, TypeRepMapBlob:
				target := reflect.New(field.Type).Interface()
				if err := gob.NewDecoder(bytes.NewReader(raw.([]byte))).Decode(target); err != nil {
					return nil, fmt.Errorf("gob decode %s: %w", field.Opts.KeyName, err)
				}
				elem.FieldByName(field.RealName).Set(reflect.ValueOf(target).Elem())
			case TypeRepFloat:
				elem.FieldByName(field.RealName).SetFloat(raw.(float64))
			default:
				return nil, fmt.Errorf("unsupported type for field %s", field.Opts.KeyName)
			}
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error for %s: %w", r.Name, err)
	}

	return results, nil
}

func (r *RegisteredStruct[T]) SelectAll() ([]*T, error) {
	return r.selectAll(r.selectAllSQL)
}

func (r *RegisteredStruct[T]) SelectAllWithFilter(filter *Filter) ([]*T, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	if err := filter.validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	fmt.Println("selecting all with filter:", fmt.Sprintf("%s WHERE %s;", r.selectAllSQL[:len(r.selectAllSQL)-1], strings.TrimSpace(filter.filter)))

	return r.selectAll(fmt.Sprintf("%s WHERE %s;", r.selectAllSQL[:len(r.selectAllSQL)-1], strings.TrimSpace(filter.filter)), filter.arguments...)
}
