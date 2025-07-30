package gomysql

import (
	"fmt"
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

	f.whereTokens = append(f.whereTokens, fmt.Sprintf("%s %s ?", key.Opts.KeyName, op))
	f.args = append(f.args, value)
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
