package gomysql

import "fmt"

func NewFilter() *Filter {
	return &Filter{
		filter:        "",
		arguments:     make([]any, 0),
		lastWasJoiner: true,
	}
}

func (f *Filter) KeyCmp(key *RegisteredStructField, op SQLOperator, value any) *Filter {
	if key == nil {
		panic("KeyCmp requires a valid key")
	}

	if !f.lastWasJoiner {
		panic("KeyCmp must be preceded by a condition join operation or nothing")
	}

	f.filter += fmt.Sprintf(" %s %s ?", key.Opts.KeyName, op)
	f.arguments = append(f.arguments, value)
	f.lastWasJoiner = false
	return f
}

func (f *Filter) And() *Filter {
	if f.lastWasJoiner {
		panic("And must be preceded by a condition")
	}

	f.filter += " AND"
	f.lastWasJoiner = true
	return f
}

func (f *Filter) Or() *Filter {
	if f.lastWasJoiner {
		panic("Or must be preceded by a condition")
	}

	f.filter += " OR"
	f.lastWasJoiner = true
	return f
}

func (f *Filter) Ordering(field *RegisteredStructField, asc bool) *Filter {
	if field == nil {
		panic("Ordering requires a valid field")
	}

	if !f.lastWasJoiner {
		panic("Ordering must be preceded by a join operation or nothing")
	}

	order := "ASC"
	if !asc {
		order = "DESC"
	}

	f.filter += fmt.Sprintf(" ORDER BY %s %s", field.RealName, order)
	f.lastWasJoiner = false
	return f
}

func (f *Filter) Limit(limit int) *Filter {
	if f.lastWasJoiner {
		panic("Limit must be preceded by a condition")
	}

	f.filter += fmt.Sprintf(" LIMIT %d", limit)
	f.lastWasJoiner = true
	return f
}

func (f *Filter) Offset(offset int) *Filter {
	if f.lastWasJoiner {
		panic("Offset must be preceded by a condition")
	}

	f.filter += fmt.Sprintf(" OFFSET %d", offset)
	f.lastWasJoiner = true
	return f
}

func (f *Filter) validate() error {
	if f.lastWasJoiner {
		return fmt.Errorf("filter ends with a joiner, expected a condition")
	}

	return nil
}
