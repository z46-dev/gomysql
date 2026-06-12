package gosqlite

import "fmt"

func (r *RegisteredStruct[T]) List() ([]any, error) {
	if r.driver == nil {
		return nil, ErrDatabaseNotInitialized
	}

	r.driver.lock.Lock()
	defer r.driver.lock.Unlock()

	rows, err := r.driver.db.Query(r.listSQL)
	if err != nil {
		return nil, fmt.Errorf("list fail %s: %w", r.Name, err)
	}
	defer rows.Close()

	var keys []any
	for rows.Next() {
		var key any
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("scan fail %s: %w", r.Name, err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}
