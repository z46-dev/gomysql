package gomysql

import (
	"fmt"
	"reflect"
)

func resolveInternalType(t reflect.Type) (TypeRepresentation, error) {
	if t.Kind() == reflect.Pointer {
		if t.Elem().Kind() == reflect.Pointer {
			return 0, fmt.Errorf("unsupported nested pointer type %s", t)
		}
		return resolveInternalType(t.Elem())
	}

	if t == timeType {
		return TypeRepTime, nil
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return TypeRepInt, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return TypeRepUint, nil
	case reflect.String:
		return TypeRepString, nil
	case reflect.Bool:
		return TypeRepBool, nil
	case reflect.Array, reflect.Slice:
		return TypeRepArrayBlob, nil
	case reflect.Struct:
		return TypeRepStructBlob, nil
	case reflect.Float32, reflect.Float64:
		return TypeRepFloat, nil
	case reflect.Map:
		return TypeRepMapBlob, nil
	default:
		return 0, fmt.Errorf("unsupported type %s", t.Kind())
	}
}

func Register[T any](structInstance T) (registered *RegisteredStruct[T], err error) {
	var structType reflect.Type = reflect.TypeOf(structInstance)

	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		err = fmt.Errorf("expected a struct, got %s", structType.Kind())
		return
	}

	registered = &RegisteredStruct[T]{
		db:     DB,
		Name:   structType.Name(),
		Type:   structType,
		Fields: make([]RegisteredStructField, 0),
	}

	for i := range structType.NumField() {
		var field reflect.StructField = structType.Field(i)
		if tag, ok := field.Tag.Lookup("gomysql"); ok {
			var opts = mustParseTag(tag)

			internalType, err := resolveInternalType(field.Type)
			if err != nil {
				return nil, fmt.Errorf("%w for field %s", err, field.Name)
			}

			registered.Fields = append(registered.Fields, RegisteredStructField{
				Opts:         opts,
				RealName:     field.Name,
				Type:         field.Type,
				Index:        field.Index,
				InternalType: internalType,
			})
		}
	}

	// Multiple primary keys or no primary key is not allowed
	var primaryKeyCount int
	for _, field := range registered.Fields {
		if field.Opts.PrimaryKey {
			primaryKeyCount++
		}
	}

	if primaryKeyCount > 1 {
		err = fmt.Errorf("multiple primary keys are not allowed in struct %s", structType.Name())
	} else if primaryKeyCount == 0 {
		err = fmt.Errorf("no primary key defined in struct %s", structType.Name())
	}

	if err != nil {
		return nil, err
	}

	generateSQLStatements(registered)

	if err = registered.runCreation(); err != nil {
		return nil, err
	}

	return
}
