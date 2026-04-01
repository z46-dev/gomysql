package gomysql

import (
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
	values := make([]any, len(r.nonInsertionOrdered)+1)
	scanArgs := make([]any, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		item := new(T)

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan all from %s: %w", r.Name, err)
		}

		elem := reflect.ValueOf(item).Elem()
		pkValue := values[0]
		if pkValue == nil {
			return nil, fmt.Errorf("primary key value is nil for %s", r.Name)
		}

		pkField := elem.FieldByIndex(r.PrimaryKeyField.Index)
		if err := assignDecodedValue(pkField, r.PrimaryKeyField, pkValue); err != nil {
			return nil, err
		}

		for i, field := range r.nonInsertionOrdered {
			raw := values[i+1]
			fieldValue := elem.FieldByIndex(field.Index)
			if err := assignDecodedValue(fieldValue, field, raw); err != nil {
				return nil, err
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

	var (
		filterString string
		filterArgs   []any
		err          error
	)

	if filterString, filterArgs, err = filter.Build(); err != nil {
		return nil, fmt.Errorf("failed to build filter: %w", err)
	}

	return r.selectAll(fmt.Sprintf("%s %s;", r.selectAllSQL[:len(r.selectAllSQL)-1], strings.TrimSpace(filterString)), filterArgs...)
}
