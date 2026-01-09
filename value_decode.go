package gomysql

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// sqlBool normalizes various SQL bool representations (int, string, []byte) to a Go bool.
func sqlBool(raw any) (bool, error) {
	switch v := raw.(type) {
	case nil:
		return false, nil
	case bool:
		return v, nil
	case int64:
		return v != 0, nil
	case int:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case []byte:
		return parseBoolString(string(v))
	case string:
		return parseBoolString(v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", raw)
	}
}

func parseBoolString(s string) (bool, error) {
	cleaned := strings.TrimSpace(strings.ToLower(s))
	switch cleaned {
	case "1", "true":
		return true, nil
	case "0", "false", "":
		return false, nil
	default:
		if b, err := strconv.ParseBool(cleaned); err == nil {
			return b, nil
		}
		return false, fmt.Errorf("cannot parse %q as bool", s)
	}
}

func decodeSQLValue(field RegisteredStructField, raw any) (any, error) {
	switch field.InternalType {
	case TypeRepInt:
		return raw.(int64), nil
	case TypeRepUint:
		return raw.(uint64), nil
	case TypeRepString:
		return raw.(string), nil
	case TypeRepBool:
		boolean, err := sqlBool(raw)
		if err != nil {
			return nil, fmt.Errorf("convert %s to bool: %w", field.Opts.KeyName, err)
		}
		return boolean, nil
	case TypeRepArrayBlob, TypeRepStructBlob, TypeRepMapBlob:
		if raw == nil {
			if field.InternalType == TypeRepArrayBlob && field.Type.Kind() == reflect.Slice {
				return reflect.Zero(field.Type).Interface(), nil
			}
			if field.InternalType == TypeRepStructBlob && field.Type == timeType {
				return time.Time{}, nil
			}
		}
		if field.InternalType == TypeRepArrayBlob && field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String {
			bytesRaw, ok := raw.([]byte)
			if !ok {
				return nil, fmt.Errorf("unsupported array blob type %T for field %s", raw, field.Opts.KeyName)
			}
			if decoded, handled, err := decodeStringSlice(bytesRaw); err != nil {
				return nil, fmt.Errorf("decode string slice %s: %w", field.Opts.KeyName, err)
			} else if handled {
				return decoded, nil
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
				return decoded, nil
			}
		}
		target := reflect.New(field.Type).Interface()
		if err := gob.NewDecoder(bytes.NewReader(raw.([]byte))).Decode(target); err != nil {
			return nil, fmt.Errorf("gob decode %s: %w", field.Opts.KeyName, err)
		}
		return reflect.ValueOf(target).Elem().Interface(), nil
	case TypeRepFloat:
		return raw.(float64), nil
	case TypeRepPointer:
		if raw == nil {
			return nil, nil
		}
		if field.Type.Elem() == timeType {
			bytesRaw, ok := raw.([]byte)
			if !ok {
				return nil, fmt.Errorf("unsupported time blob type %T for field %s", raw, field.Opts.KeyName)
			}
			if decoded, handled, err := decodeTimeValue(bytesRaw); err != nil {
				return nil, fmt.Errorf("decode time %s: %w", field.Opts.KeyName, err)
			} else if handled {
				return &decoded, nil
			}
		}
		target := reflect.New(field.Type.Elem())
		if err := gob.NewDecoder(bytes.NewReader(raw.([]byte))).Decode(target.Interface()); err != nil {
			return nil, fmt.Errorf("gob decode %s: %w", field.Opts.KeyName, err)
		}
		return target.Interface(), nil
	default:
		return nil, fmt.Errorf("unsupported type for field %s", field.Opts.KeyName)
	}
}
