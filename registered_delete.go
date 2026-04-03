package gomysql

import "fmt"

func (r *RegisteredStruct[T]) Delete(primaryKeyValue any) error {
	if r.db == nil {
		return ErrDatabaseNotInitialized
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	if _, err := r.db.db.Exec(r.deleteSQL, primaryKeyValue); err != nil {
		return fmt.Errorf("delete fail %s: %w", r.Name, err)
	}

	return nil
}

func (r *RegisteredStruct[T]) DeleteWithFilter(filter *Filter) (int64, error) {
	if r.db == nil {
		return 0, ErrDatabaseNotInitialized
	}

	sql, args, err := r.buildDeleteWithFilterSQL(filter)
	if err != nil {
		return 0, err
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	result, err := r.db.db.Exec(sql, args...)
	if err != nil {
		return 0, fmt.Errorf("delete with filter fail %s: %w", r.Name, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete with filter rows affected %s: %w", r.Name, err)
	}

	return rows, nil
}

func (r *RegisteredStruct[T]) buildDeleteWithFilterSQL(filter *Filter) (string, []any, error) {
	switch {
	case filterHasSelectionModifiers(filter):
		filterClause, filterArgs, err := buildFilterClause(filter)
		if err != nil {
			return "", nil, err
		}

		pk := r.PrimaryKeyField.Opts.KeyName
		sql := fmt.Sprintf("DELETE FROM %s WHERE %s IN (SELECT %s FROM %s", r.Name, pk, pk, r.Name)
		if filterClause != "" {
			sql += " " + filterClause
		}
		sql += ");"
		return sql, filterArgs, nil
	case filterHasWhere(filter):
		whereClause, whereArgs, err := buildWhereClause(filter)
		if err != nil {
			return "", nil, err
		}

		return fmt.Sprintf("DELETE FROM %s %s;", r.Name, whereClause), whereArgs, nil
	default:
		return fmt.Sprintf("DELETE FROM %s;", r.Name), nil, nil
	}
}
