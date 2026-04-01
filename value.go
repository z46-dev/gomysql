package gomysql

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"time"
)

func baseTypeOf(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func derefValue(v reflect.Value) (reflect.Value, bool) {
	if v.Kind() != reflect.Pointer {
		return v, true
	}
	if v.IsNil() {
		return reflect.Value{}, false
	}
	return v.Elem(), true
}

func getSQLValueOf(field RegisteredStructField, structField reflect.Value) (any, error) {
	value, ok := derefValue(structField)
	if !ok {
		return nil, nil
	}

	switch field.InternalType {
	case TypeRepInt:
		return value.Int(), nil
	case TypeRepString:
		return value.String(), nil
	case TypeRepBool:
		return value.Bool(), nil
	case TypeRepArrayBlob, TypeRepStructBlob, TypeRepMapBlob:
		fieldBaseType := baseTypeOf(field.Type)
		if field.InternalType == TypeRepArrayBlob && fieldBaseType.Kind() == reflect.Slice && fieldBaseType.Elem().Kind() == reflect.String {
			return encodeStringSlice(value.Interface().([]string)), nil
		}
		if field.InternalType == TypeRepStructBlob && fieldBaseType == timeType {
			encoded, err := encodeTimeValue(value.Interface().(time.Time))
			if err != nil {
				return nil, fmt.Errorf("encode time %s: %w", field.Opts.KeyName, err)
			}
			return encoded, nil
		}
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(value.Interface()); err != nil {
			return nil, fmt.Errorf("gob encoding %s: %w", field.Opts.KeyName, err)
		}

		return buf.Bytes(), nil
	case TypeRepFloat:
		return value.Float(), nil
	case TypeRepUint:
		return value.Uint(), nil
	default:
		return nil, fmt.Errorf("unsupported type for field %s", field.Opts.KeyName)
	}
}
