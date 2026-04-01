package gomysql

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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

func sqlInt64(raw any) (int64, error) {
	switch v := raw.(type) {
	case nil:
		return 0, nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", raw)
	}
}

func sqlUint64(raw any) (uint64, error) {
	switch v := raw.(type) {
	case nil:
		return 0, nil
	case uint64:
		return v, nil
	case uint:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case []byte:
		return strconv.ParseUint(string(v), 10, 64)
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", raw)
	}
}

func sqlFloat64(raw any) (float64, error) {
	switch v := raw.(type) {
	case nil:
		return 0, nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", raw)
	}
}

func sqlString(raw any) (string, error) {
	switch v := raw.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("cannot convert %T to string", raw)
	}
}

func assignDecodedValue(fieldValue reflect.Value, field RegisteredStructField, raw any) error {
	if fieldValue.Kind() == reflect.Pointer {
		if raw == nil {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
			return nil
		}

		target := reflect.New(fieldValue.Type().Elem())
		if err := assignDecodedValue(target.Elem(), RegisteredStructField{
			Opts:         field.Opts,
			RealName:     field.RealName,
			Type:         fieldValue.Type().Elem(),
			Index:        field.Index,
			InternalType: field.InternalType,
		}, raw); err != nil {
			return err
		}
		fieldValue.Set(target)
		return nil
	}

	switch field.InternalType {
	case TypeRepInt:
		value, err := sqlInt64(raw)
		if err != nil {
			return fmt.Errorf("convert %s to int: %w", field.Opts.KeyName, err)
		}
		fieldValue.SetInt(value)
	case TypeRepUint:
		value, err := sqlUint64(raw)
		if err != nil {
			return fmt.Errorf("convert %s to uint: %w", field.Opts.KeyName, err)
		}
		fieldValue.SetUint(value)
	case TypeRepString:
		value, err := sqlString(raw)
		if err != nil {
			return fmt.Errorf("convert %s to string: %w", field.Opts.KeyName, err)
		}
		fieldValue.SetString(value)
	case TypeRepBool:
		boolean, err := sqlBool(raw)
		if err != nil {
			return fmt.Errorf("convert %s to bool: %w", field.Opts.KeyName, err)
		}
		fieldValue.SetBool(boolean)
	case TypeRepFloat:
		value, err := sqlFloat64(raw)
		if err != nil {
			return fmt.Errorf("convert %s to float: %w", field.Opts.KeyName, err)
		}
		fieldValue.SetFloat(value)
	case TypeRepArrayBlob, TypeRepStructBlob, TypeRepMapBlob:
		if raw == nil {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
			return nil
		}
		if field.InternalType == TypeRepArrayBlob && fieldValue.Kind() == reflect.Slice && fieldValue.Type().Elem().Kind() == reflect.String {
			bytesRaw, ok := raw.([]byte)
			if !ok {
				return fmt.Errorf("unsupported array blob type %T for field %s", raw, field.Opts.KeyName)
			}
			if decoded, handled, err := decodeStringSlice(bytesRaw); err != nil {
				return fmt.Errorf("decode string slice %s: %w", field.Opts.KeyName, err)
			} else if handled {
				if decoded == nil {
					fieldValue.Set(reflect.Zero(fieldValue.Type()))
					return nil
				}
				fieldValue.Set(reflect.ValueOf(decoded))
				return nil
			}
		}
		if field.InternalType == TypeRepStructBlob && fieldValue.Type() == timeType {
			bytesRaw, ok := raw.([]byte)
			if !ok {
				return fmt.Errorf("unsupported time blob type %T for field %s", raw, field.Opts.KeyName)
			}
			if decoded, handled, err := decodeTimeValue(bytesRaw); err != nil {
				return fmt.Errorf("decode time %s: %w", field.Opts.KeyName, err)
			} else if handled {
				fieldValue.Set(reflect.ValueOf(decoded))
				return nil
			}
		}
		if err := gob.NewDecoder(bytes.NewReader(raw.([]byte))).Decode(fieldValue.Addr().Interface()); err != nil {
			return fmt.Errorf("gob decode %s: %w", field.Opts.KeyName, err)
		}
	default:
		return fmt.Errorf("unsupported type for field %s", field.Opts.KeyName)
	}

	return nil
}

func decodeSQLValue(field RegisteredStructField, raw any) (any, error) {
	target := reflect.New(field.Type).Elem()
	if err := assignDecodedValue(target, field, raw); err != nil {
		return nil, err
	}
	return target.Interface(), nil
}
