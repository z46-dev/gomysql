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
		fieldValue := elem.FieldByIndex(field.Index)

		switch field.InternalType {
		case TypeRepInt:
			if raw == nil {
				fieldValue.SetInt(0)
				continue
			}
			fieldValue.SetInt(raw.(int64))
		case TypeRepUint:
			if raw == nil {
				fieldValue.SetUint(0)
				continue
			}
			fieldValue.SetUint(raw.(uint64))
		case TypeRepString:
			if raw == nil {
				fieldValue.SetString("")
				continue
			}
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
			if raw == nil {
				fieldValue.SetFloat(0)
				continue
			}
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

	if primaryKeyField := r.PrimaryKeyField; primaryKeyField.RealName != "" {
		elem.FieldByIndex(primaryKeyField.Index).Set(reflect.ValueOf(primaryKeyValue))
	}

	return item, nil
}
