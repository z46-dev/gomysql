package gomysql

import "fmt"

func NewFilter() *Filter {
	return &Filter{
		filter:        "",
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

	f.filter += fmt.Sprintf(" %s %s %s", key.Opts.KeyName, op, value)
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

	if f.lastWasJoiner {
		panic("Ordering must be preceded by a condition")
	}

	order := "ASC"
	if !asc {
		order = "DESC"
	}

	f.filter += fmt.Sprintf(" ORDER BY %s %s", field.RealName, order)
	f.lastWasJoiner = true
	return f
}

func (f *Filter) validate() error {
	if f.lastWasJoiner {
		return fmt.Errorf("filter ends with a joiner, expected a condition")
	}

	return nil
}
