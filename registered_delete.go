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
