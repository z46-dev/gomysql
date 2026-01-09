package gomysql

import (
	"fmt"
	"strings"
)

func SetField(field *RegisteredStructField, value any) UpdateAssignment {
	if field == nil {
		panic("SetField requires a valid field")
	}

	return UpdateAssignment{
		clause: fmt.Sprintf("%s = ?", field.Opts.KeyName),
		args:   []any{value},
	}
}

func SetExpr(field *RegisteredStructField, expr string, args ...any) UpdateAssignment {
	if field == nil {
		panic("SetExpr requires a valid field")
	}

	if strings.TrimSpace(expr) == "" {
		panic("SetExpr requires a non-empty expression")
	}

	return UpdateAssignment{
		clause: fmt.Sprintf("%s = %s", field.Opts.KeyName, expr),
		args:   args,
	}
}

func SetAdd(field *RegisteredStructField, value any) UpdateAssignment {
	return SetExpr(field, fmt.Sprintf("%s + ?", field.Opts.KeyName), value)
}

func SetSub(field *RegisteredStructField, value any) UpdateAssignment {
	return SetExpr(field, fmt.Sprintf("%s - ?", field.Opts.KeyName), value)
}

func SetMul(field *RegisteredStructField, value any) UpdateAssignment {
	return SetExpr(field, fmt.Sprintf("%s * ?", field.Opts.KeyName), value)
}

func SetDiv(field *RegisteredStructField, value any) UpdateAssignment {
	return SetExpr(field, fmt.Sprintf("%s / ?", field.Opts.KeyName), value)
}

func (r *RegisteredStruct[T]) UpdateWithFilter(filter *Filter, assignments ...UpdateAssignment) (int64, error) {
	if r.db == nil {
		return 0, ErrDatabaseNotInitialized
	}

	setClause, setArgs, err := buildUpdateAssignments(assignments)
	if err != nil {
		return 0, err
	}

	filterClause, filterArgs, err := buildFilterClause(filter)
	if err != nil {
		return 0, err
	}

	sql := fmt.Sprintf("UPDATE %s SET %s", r.Name, setClause)
	if filterClause != "" {
		sql += " " + filterClause
	}
	sql += ";"

	args := append(setArgs, filterArgs...)

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	result, err := r.db.db.Exec(sql, args...)
	if err != nil {
		return 0, fmt.Errorf("update with filter fail %s: %w", r.Name, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("update with filter rows affected %s: %w", r.Name, err)
	}

	return rows, nil
}

func (r *RegisteredStruct[T]) UpdateWithFilterReturning(filter *Filter, returning []*RegisteredStructField, assignments ...UpdateAssignment) ([]ReturnedValues, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	if len(returning) == 0 {
		return nil, fmt.Errorf("returning requires at least one field")
	}

	setClause, setArgs, err := buildUpdateAssignments(assignments)
	if err != nil {
		return nil, err
	}

	filterClause, filterArgs, err := buildFilterClause(filter)
	if err != nil {
		return nil, err
	}

	var returningCols []string
	for _, field := range returning {
		if field == nil {
			return nil, fmt.Errorf("returning requires valid fields")
		}
		returningCols = append(returningCols, field.Opts.KeyName)
	}

	sql := fmt.Sprintf("UPDATE %s SET %s", r.Name, setClause)
	if filterClause != "" {
		sql += " " + filterClause
	}
	sql += " RETURNING " + strings.Join(returningCols, ", ") + ";"

	args := append(setArgs, filterArgs...)

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	rows, err := r.db.db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("update returning fail %s: %w", r.Name, err)
	}
	defer rows.Close()

	var results []ReturnedValues
	values := make([]any, len(returning))
	scanArgs := make([]any, len(returning))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("update returning scan fail %s: %w", r.Name, err)
		}

		row := make(ReturnedValues, len(returning))
		for i, field := range returning {
			decoded, err := decodeSQLValue(*field, values[i])
			if err != nil {
				return nil, fmt.Errorf("update returning decode %s: %w", field.Opts.KeyName, err)
			}
			row[field.Opts.KeyName] = decoded
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("update returning rows error %s: %w", r.Name, err)
	}

	return results, nil
}

func buildUpdateAssignments(assignments []UpdateAssignment) (string, []any, error) {
	if len(assignments) == 0 {
		return "", nil, fmt.Errorf("update requires at least one assignment")
	}

	var clauses []string
	var args []any
	for _, assignment := range assignments {
		if strings.TrimSpace(assignment.clause) == "" {
			return "", nil, fmt.Errorf("update requires non-empty assignments")
		}
		clauses = append(clauses, assignment.clause)
		if len(assignment.args) > 0 {
			args = append(args, assignment.args...)
		}
	}

	return strings.Join(clauses, ", "), args, nil
}

func buildFilterClause(filter *Filter) (string, []any, error) {
	if filter == nil {
		return "", nil, nil
	}

	filterString, filterArgs, err := filter.Build()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build filter: %w", err)
	}

	filterString = strings.TrimSpace(filterString)
	return filterString, filterArgs, nil
}
