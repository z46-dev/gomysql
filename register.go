package gomysql

import (
	"fmt"
	"reflect"
)

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

			var internalType TypeRepresentation
			switch field.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				internalType = TypeRepInt
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				internalType = TypeRepUint
			case reflect.String:
				internalType = TypeRepString
			case reflect.Bool:
				internalType = TypeRepBool
			case reflect.Array, reflect.Slice:
				internalType = TypeRepArrayBlob
			case reflect.Struct:
				internalType = TypeRepStructBlob
			case reflect.Float32, reflect.Float64:
				internalType = TypeRepFloat
			case reflect.Map:
				internalType = TypeRepMapBlob
			case reflect.Pointer:
				if field.Type.Elem().Kind() == reflect.Struct {
					internalType = TypeRepPointer
				} else {
					return nil, fmt.Errorf("unsupported pointer type %s for field %s", field.Type.Elem().Kind(), field.Name)
				}
			default:
				return nil, fmt.Errorf("unsupported type %s for field %s", field.Type.Kind(), field.Name)
			}

			registered.Fields = append(registered.Fields, RegisteredStructField{
				Opts:         opts,
				RealName:     field.Name,
				Type:         field.Type,
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
