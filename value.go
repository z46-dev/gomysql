package gomysql

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
)

func getSQLValueOf(field RegisteredStructField, structField reflect.Value) (any, error) {
	switch field.InternalType {
	case TypeRepInt:
		return structField.Int(), nil
	case TypeRepString:
		return structField.String(), nil
	case TypeRepBool:
		return structField.Bool(), nil
	case TypeRepArrayBlob, TypeRepStructBlob, TypeRepMapBlob:
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(structField.Interface()); err != nil {
			return nil, fmt.Errorf("gob encoding %s: %w", field.Opts.KeyName, err)
		}

		return buf.Bytes(), nil
	case TypeRepFloat:
		return structField.Float(), nil
	case TypeRepUint:
		return structField.Uint(), nil
	case TypeRepPointer:
		if structField.IsNil() {
			return nil, nil
		}
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(structField.Interface()); err != nil {
			return nil, fmt.Errorf("gob encoding %s: %w", field.Opts.KeyName, err)
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported type for field %s", field.Opts.KeyName)
	}
}
