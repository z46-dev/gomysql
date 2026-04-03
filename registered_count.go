package gomysql

import "fmt"

func (r *RegisteredStruct[T]) Count() (int64, error) {
	return r.CountWithFilter(nil)
}

func (r *RegisteredStruct[T]) CountWithFilter(filter *Filter) (int64, error) {
	if r.db == nil {
		return 0, ErrDatabaseNotInitialized
	}

	sql, args, err := r.buildCountSQL(filter)
	if err != nil {
		return 0, err
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	var count int64
	if err := r.db.db.QueryRow(sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count fail %s: %w", r.Name, err)
	}

	return count, nil
}

func (r *RegisteredStruct[T]) buildCountSQL(filter *Filter) (string, []any, error) {
	switch {
	case filterHasSelectionModifiers(filter):
		filterClause, filterArgs, err := buildFilterClause(filter)
		if err != nil {
			return "", nil, err
		}

		sql := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT %s FROM %s", r.PrimaryKeyField.Opts.KeyName, r.Name)
		if filterClause != "" {
			sql += " " + filterClause
		}
		sql += ") AS filtered_rows;"
		return sql, filterArgs, nil
	case filterHasWhere(filter):
		whereClause, whereArgs, err := buildWhereClause(filter)
		if err != nil {
			return "", nil, err
		}

		return fmt.Sprintf("SELECT COUNT(*) FROM %s %s;", r.Name, whereClause), whereArgs, nil
	default:
		return fmt.Sprintf("SELECT COUNT(*) FROM %s;", r.Name), nil, nil
	}
}
