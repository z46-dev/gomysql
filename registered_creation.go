package gomysql

func (r *RegisteredStruct[T]) runCreation() (err error) {
	if r.db == nil {
		err = ErrDatabaseNotInitialized
		return
	}

	r.db.lock.Lock()
	defer r.db.lock.Unlock()

	if _, err = r.db.db.Exec(r.createTableSQL); err != nil {
		return
	}

	return
}
