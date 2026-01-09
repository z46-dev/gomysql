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
			fieldValue := elem.FieldByIndex(field.Index)
			switch field.InternalType {
			case TypeRepInt:
				fieldValue.SetInt(raw.(int64))
			case TypeRepUint:
				fieldValue.SetUint(raw.(uint64))
			case TypeRepString:
				fieldValue.SetString(raw.(string))
			case TypeRepBool:
				boolean, err := sqlBool(raw)
				if err != nil {
					return nil, fmt.Errorf("convert %s to bool: %w", field.Opts.KeyName, err)
				}
				fieldValue.SetBool(boolean)
			case TypeRepArrayBlob, TypeRepStructBlob, TypeRepMapBlob:
				if raw == nil {
					fieldValue.Set(reflect.Zero(field.Type))
					continue
				}
				if field.InternalType == TypeRepArrayBlob && field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String {
					bytesRaw, ok := raw.([]byte)
					if !ok {
						return nil, fmt.Errorf("unsupported array blob type %T for field %s", raw, field.Opts.KeyName)
					}
					if decoded, handled, err := decodeStringSlice(bytesRaw); err != nil {
						return nil, fmt.Errorf("decode string slice %s: %w", field.Opts.KeyName, err)
					} else if handled {
						if decoded == nil {
							fieldValue.Set(reflect.Zero(field.Type))
							continue
						}
						fieldValue.Set(reflect.ValueOf(decoded))
						continue
					}
				}
				if field.InternalType == TypeRepStructBlob && field.Type == timeType {
					bytesRaw, ok := raw.([]byte)
					if !ok {
						return nil, fmt.Errorf("unsupported time blob type %T for field %s", raw, field.Opts.KeyName)
					}
					if decoded, handled, err := decodeTimeValue(bytesRaw); err != nil {
						return nil, fmt.Errorf("decode time %s: %w", field.Opts.KeyName, err)
					} else if handled {
						fieldValue.Set(reflect.ValueOf(decoded))
						continue
					}
				}
				target := reflect.New(field.Type).Interface()
				if err := gob.NewDecoder(bytes.NewReader(raw.([]byte))).Decode(target); err != nil {
					return nil, fmt.Errorf("gob decode %s: %w", field.Opts.KeyName, err)
				}
				fieldValue.Set(reflect.ValueOf(target).Elem())
			case TypeRepFloat:
				fieldValue.SetFloat(raw.(float64))
			case TypeRepPointer:
				if raw == nil {
					fieldValue.Set(reflect.Zero(field.Type))
					continue
				}
				if field.Type.Elem() == timeType {
					bytesRaw, ok := raw.([]byte)
					if !ok {
						return nil, fmt.Errorf("unsupported time blob type %T for field %s", raw, field.Opts.KeyName)
					}
					if decoded, handled, err := decodeTimeValue(bytesRaw); err != nil {
						return nil, fmt.Errorf("decode time %s: %w", field.Opts.KeyName, err)
					} else if handled {
						fieldValue.Set(reflect.ValueOf(&decoded))
						continue
					}
				}
				target := reflect.New(field.Type.Elem())
				if err := gob.NewDecoder(bytes.NewReader(raw.([]byte))).Decode(target.Interface()); err != nil {
					return nil, fmt.Errorf("gob decode %s: %w", field.Opts.KeyName, err)
				}
				fieldValue.Set(target)
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
