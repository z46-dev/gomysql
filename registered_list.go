package gomysql

import "fmt"

func (r *RegisteredStruct[T]) List() ([]any, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	rows, err := r.db.db.Query(r.listSQL)
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
