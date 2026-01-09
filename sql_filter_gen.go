package gomysql

import (
	"fmt"
	"reflect"
	"strings"
)

func NewFilter() *Filter {
	return &Filter{
		whereTokens:   []string{},
		args:          []any{},
		lastWasJoiner: true,
	}
}

func (f *Filter) KeyCmp(key *RegisteredStructField, op SQLOperator, value any) *Filter {
	if key == nil {
		panic("KeyCmp requires a valid key")
	}

	if !f.lastWasJoiner {
		panic("KeyCmp must be preceded by a joiner (And/Or) or be the first condition")
	}

	switch op {
	case OpIsNull, OpIsNotNull:
		if value != nil {
			panic("KeyCmp with IS NULL/IS NOT NULL does not accept a value")
		}
		f.whereTokens = append(f.whereTokens, fmt.Sprintf("%s %s", key.Opts.KeyName, op))
	case OpIn, OpNotIn:
		if value == nil {
			panic("KeyCmp with IN/NOT IN requires a slice or array value")
		}
		val := reflect.ValueOf(value)
		if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
			panic("KeyCmp with IN/NOT IN requires a slice or array value")
		}
		if val.Len() == 0 {
			panic("KeyCmp with IN/NOT IN requires at least one value")
		}
		placeholders := strings.Repeat("?, ", val.Len()-1) + "?"
		f.whereTokens = append(f.whereTokens, fmt.Sprintf("%s %s (%s)", key.Opts.KeyName, op, placeholders))
		for i := 0; i < val.Len(); i++ {
			f.args = append(f.args, val.Index(i).Interface())
		}
	default:
		f.whereTokens = append(f.whereTokens, fmt.Sprintf("%s %s ?", key.Opts.KeyName, op))
		f.args = append(f.args, value)
	}
	f.lastWasJoiner = false
	return f
}

func (f *Filter) And() *Filter {
	if f.lastWasJoiner {
		panic("And must be preceded by a condition")
	}

	f.whereTokens = append(f.whereTokens, "AND")
	f.lastWasJoiner = true
	return f
}

func (f *Filter) Or() *Filter {
	if f.lastWasJoiner {
		panic("Or must be preceded by a condition")
	}

	f.whereTokens = append(f.whereTokens, "OR")
	f.lastWasJoiner = true
	return f
}

func (f *Filter) OpenGroup() *Filter {
	if !f.lastWasJoiner {
		panic("OpenGroup must be preceded by a joiner or be the first condition")
	}

	f.whereTokens = append(f.whereTokens, "(")
	f.lastWasJoiner = true
	return f
}

func (f *Filter) CloseGroup() *Filter {
	if f.lastWasJoiner {
		panic("CloseGroup must be preceded by a condition")
	}

	f.whereTokens = append(f.whereTokens, ")")
	f.lastWasJoiner = false
	return f
}

func (f *Filter) Ordering(field *RegisteredStructField, asc bool) *Filter {
	if field == nil {
		panic("Ordering requires a valid field")
	}

	var dir string = "ASC"
	if !asc {
		dir = "DESC"
	}

	f.orderByClause = fmt.Sprintf("ORDER BY %s %s", field.RealName, dir)
	f.lastWasJoiner = false
	return f
}

func (f *Filter) Limit(n int) *Filter {
	if f.lastWasJoiner && len(f.whereTokens) > 0 {
		panic("Limit must be preceded by a condition")
	}

	f.limitClause = fmt.Sprintf("LIMIT %d", n)
	f.lastWasJoiner = false
	return f
}

func (f *Filter) Offset(n int) *Filter {
	if f.lastWasJoiner {
		panic("Offset must be preceded by a condition or LIMIT")
	}

	f.offsetClause = fmt.Sprintf("OFFSET %d", n)
	f.lastWasJoiner = true
	return f
}

func (f *Filter) Build() (sqlFragment string, args []any, err error) {
	if f.lastWasJoiner && len(f.whereTokens) > 0 {
		return "", nil, fmt.Errorf("filter ends with a joiner; expected a condition")
	}

	var parts []string
	if len(f.whereTokens) > 0 {
		parts = append(parts, "WHERE "+strings.Join(f.whereTokens, " "))
	}

	if f.orderByClause != "" {
		parts = append(parts, f.orderByClause)
	}

	if f.limitClause != "" {
		parts = append(parts, f.limitClause)
	}

	if f.offsetClause != "" {
		parts = append(parts, f.offsetClause)
	}

	return strings.Join(parts, " "), f.args, nil
}
